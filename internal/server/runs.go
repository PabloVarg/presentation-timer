package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PabloVarg/presentation-timer/internal/helpers"
	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

func RunPresentation(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	runs := make(map[int64]RunTask)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := uuid.NewRandom()
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}

		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		if _, ok := runs[ID]; !ok {
			runs[ID] = NewRun(ID, logger, queriesStore)
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
		defer func() {
			conn.Close()
			runs[ID].RemoveConnection(u.String())

			if runs[ID].Terminated() {
				delete(runs, ID)
			}
		}()

		runs[ID].AddConnection(u.String(), conn)

		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				return
			}

			fields := strings.Fields(string(p))
			switch {
			case fields[0] == "start":
				runs[ID].SendMsg(StartPresentation, WithConn(conn))
			case fields[0] == "pause":
				runs[ID].SendMsg(PausePresentation, WithConn(conn))
			case fields[0] == "resume":
				runs[ID].SendMsg(ResumePresentation, WithConn(conn))
			case fields[0] == "step":
				if len(fields) != 2 {
					continue
				}

				step, err := strconv.Atoi(fields[1])
				if err != nil {
					continue
				}

				runs[ID].SendMsg(StepInto, WithStep(int32(step)), WithConn(conn))
			default:
				conn.WriteJSON(map[string]any{
					"error": "command not recognized",
				})
			}

		}
	})
}

type RunTask struct {
	presentationID int64
	logger         *slog.Logger
	conns          map[string]*websocket.Conn
	queriesStore   *queries.Queries
	// sections
	sections []queries.Section
	// runs state
	ctx            context.Context
	cancel         context.CancelFunc
	timer          *time.Timer
	timerEnd       time.Time // Stores end of timer, for pause events
	timerRemaining time.Duration
	step           int32
	msg            chan TaskMsg
}

type TaskMsg struct {
	conn       *websocket.Conn
	action     int
	targetStep int32
}

const (
	StartPresentation  = iota
	PausePresentation  = iota
	ResumePresentation = iota
	StepInto
)

func NewRun(
	presentationID int64,
	logger *slog.Logger,
	queriesStore *queries.Queries,
) RunTask {
	stoppedTimer := time.NewTimer(0)
	stoppedTimer.Stop()

	ctx, cancel := context.WithCancel(context.Background())

	task := RunTask{
		presentationID: presentationID,
		logger:         logger,
		queriesStore:   queriesStore,
		conns:          make(map[string]*websocket.Conn),
		// runs state
		ctx:    ctx,
		cancel: cancel,
		timer:  stoppedTimer,
		msg:    make(chan TaskMsg),
	}
	go task.Run()

	return task
}

func (t RunTask) AddConnection(ID string, conn *websocket.Conn) {
	t.conns[ID] = conn
}

func (t RunTask) RemoveConnection(ID string) {
	delete(t.conns, ID)

	if len(t.conns) == 0 && t.ctx != nil {
		select {
		case <-t.ctx.Done():
			return
		default:
			t.cancel()
		}
	}
}

func (t RunTask) Terminated() bool {
	return len(t.conns) == 0
}

func (t *RunTask) Run() {
	t.logger.Info("pr run", "event", "coroutine started")

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-t.timer.C:
			t.logger.Info("pr run", "tick", "tock")
			t.step += 1
			if int(t.step) >= len(t.sections) {
				t.timer.Stop()
				t.Broadcast(map[string]any{
					"status": "end",
				})
				continue
			}

			t.timer = time.NewTimer(t.sections[t.step].Duration)
			t.timerEnd = time.Now().Add(t.sections[t.step].Duration)

			t.Broadcast(map[string]any{
				"step":         t.sections[t.step].Name,
				"finish_time":  t.timerEnd,
				"remaining_ms": t.sections[t.step].Duration.Milliseconds(),
			})
		case msg := <-t.msg:
			if err := t.HandleMsg(msg); err != nil {
				t.RespondToMsg(msg, map[string]any{
					"error": err.Error(),
				})
			}
		}
	}
}

func (t *RunTask) HandleMsg(msg TaskMsg) error {
	t.logger.Info("handle message", "msg", msg, "action", msg.action)

	switch msg.action {
	case StartPresentation:
		t.logger.Info("handle message", "case", "start presentation")
		dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sections, err := t.queriesStore.GetSectionsByPosition(dbCtx, t.presentationID)
		if err != nil {
			return err
		}

		t.sections = sections
		t.step = -1
		t.timer = time.NewTimer(0)
	case PausePresentation:
		t.logger.Info("handle message", "case", "pause presentation")
		t.timerRemaining = t.timerEnd.Sub(time.Now())
		t.timer.Stop()

		t.Broadcast(map[string]any{
			"step":         t.sections[t.step].Name,
			"remaining_ms": t.sections[t.step].Duration.Milliseconds(),
		})
	case ResumePresentation:
		t.logger.Info("handle message", "case", "resume presentation")
		t.timer = time.NewTimer(t.timerRemaining)
		t.timerEnd = time.Now().Add(t.timerRemaining)
	case StepInto:
		t.logger.Info("handle message", "case", "step presentation")
		if msg.targetStep < 0 || int(msg.targetStep) >= len(t.sections) {
			return fmt.Errorf("target step is not valid")
		}

		t.step = msg.targetStep - 1
		t.timer = time.NewTimer(0)
	}

	return nil
}

func (t RunTask) SendMsg(action int, opts ...func(*TaskMsg)) {
	msg := TaskMsg{
		action: action,
	}

	for _, opt := range opts {
		opt(&msg)
	}

	t.msg <- msg
}

func WithStep(step int32) func(*TaskMsg) {
	return func(tm *TaskMsg) {
		tm.targetStep = step
	}
}

func WithConn(conn *websocket.Conn) func(*TaskMsg) {
	return func(tm *TaskMsg) {
		tm.conn = conn
	}
}

func (t RunTask) Broadcast(msg any) {
	b, err := json.Marshal(msg)
	if err != nil {
		t.logger.Error("ws broadcast", "err", err)
	}

	pm, err := websocket.NewPreparedMessage(websocket.BinaryMessage, b)
	if err != nil {
		t.logger.Error("ws broadcast", "err", err)
	}

	g := new(errgroup.Group)
	for _, conn := range t.conns {
		g.Go(func() error {
			return conn.WritePreparedMessage(pm)
		})
	}
	if err := g.Wait(); err != nil {
		t.logger.Error("ws broadcast", "err", err)
	}
}

func (t RunTask) RespondToMsg(msg TaskMsg, message any) {
	if msg.conn == nil {
		return
	}

	if err := msg.conn.WriteJSON(message); err != nil {
		switch {
		case errors.Is(err, websocket.ErrCloseSent):
			t.logger.Info("send response to closed conn", "message", msg, "response", message)
			return
		default:
			t.logger.Error("failed to send ws response", "err", err)
		}
	}
}
