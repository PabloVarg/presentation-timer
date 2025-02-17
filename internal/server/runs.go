package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/PabloVarg/presentation-timer/internal/helpers"
	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

type RunStatusResponse struct {
	State  string          `json:"state"`
	Step   queries.Section `json:"step"`
	MsLeft int64           `json:"ms_left"`
	// errors
	Err string `json:"error,omitempty"`
}

func RunPresentation(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type input struct {
		Action string `json:"action"`
		Step   *int32 `json:"step"`
	}

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
			runs[ID], err = NewRun(ID, logger, queriesStore)
			if err != nil {
				helpers.InternalError(w, logger, err)
				return
			}
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

			var message input
			if err := json.NewDecoder(bytes.NewReader(p)).Decode(&message); err != nil {
				conn.WriteJSON(helpers.ErrorResponse{
					Error: err.Error(),
				})
				continue
			}

			switch message.Action {
			case "status":
				runs[ID].SendMsg(Status, WithConn(conn))
			case "start":
				runs[ID].SendMsg(StartPresentation, WithConn(conn))
			case "pause":
				runs[ID].SendMsg(PausePresentation, WithConn(conn))
			case "resume":
				runs[ID].SendMsg(ResumePresentation, WithConn(conn))
			case "step":
				if message.Step == nil {
					conn.WriteJSON(helpers.ErrorResponse{
						Error: "a step must be given for the step action",
					})
					break
				}

				runs[ID].SendMsg(StepInto, WithStep(*message.Step), WithConn(conn))
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
	// sections
	sections []queries.Section
	// runs state
	isRunning     bool
	ctx           context.Context
	cancel        context.CancelFunc
	timer         *time.Timer
	timerEnd      time.Time     // Stores end of timer, for pause events
	timeRemaining time.Duration // Stores the time remaining for next step on pause events
	step          int32
	msg           chan TaskMsg
}

type TaskMsg struct {
	conn       *websocket.Conn
	action     int
	targetStep int32
}

const (
	StartPresentation = iota
	PausePresentation
	ResumePresentation
	StepInto
	Status
)

func NewRun(
	presentationID int64,
	logger *slog.Logger,
	queriesStore *queries.Queries,
) (RunTask, error) {
	stoppedTimer := time.NewTimer(0)

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sections, err := queriesStore.GetSectionsByPosition(dbCtx, presentationID)
	if err != nil {
		return RunTask{}, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	task := RunTask{
		presentationID: presentationID,
		logger:         logger,
		conns:          make(map[string]*websocket.Conn),
		sections:       sections,
		// runs state
		isRunning: false,
		ctx:       ctx,
		cancel:    cancel,
		timer:     stoppedTimer,
		msg:       make(chan TaskMsg),
		step:      -1,
	}
	go task.Run()

	return task, nil
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
			t.timerEnd = time.Time{}

			if int(t.step) >= len(t.sections) {
				t.step -= 1
				t.timer.Stop()
				t.isRunning = false
				t.Broadcast(t.GetRunState())
				continue
			}

			t.timer = time.NewTimer(t.sections[t.step].Duration)
			t.timerEnd = time.Now().Add(t.sections[t.step].Duration)

			if t.isRunning == false {
				t.timeRemaining = t.timerEnd.Sub(time.Now())
				t.timerEnd = time.Time{}
				t.timer.Stop()
			}

			t.Broadcast(t.GetRunState())
		case msg := <-t.msg:
			if err := t.HandleMsg(msg); err != nil {
				state := t.GetRunState()
				state.Err = err.Error()

				t.RespondToMsg(msg, state)
			}
		}
	}
}

func (t *RunTask) HandleMsg(msg TaskMsg) error {
	t.logger.Info("handle message", "msg", msg, "action", msg.action)

	switch msg.action {
	case Status:
		t.Broadcast(t.GetRunState())
	case StartPresentation:
		t.logger.Info("handle message", "case", "start presentation")
		t.step = -1
		t.timer = time.NewTimer(0)
		t.isRunning = true
		t.timeRemaining = 0
	case PausePresentation:
		t.logger.Info("handle message", "case", "pause presentation")
		if t.timeRemaining == 0 {
			t.timeRemaining = t.timerEnd.Sub(time.Now())
			t.timerEnd = time.Time{}
			t.timer.Stop()
			t.isRunning = false
		}

		t.Broadcast(t.GetRunState())
	case ResumePresentation:
		t.logger.Info("handle message", "case", "resume presentation")
		if t.timeRemaining != 0 {
			t.timer = time.NewTimer(t.timeRemaining)
			t.timerEnd = time.Now().Add(t.timeRemaining)
			t.timeRemaining = 0
			t.isRunning = true
		}

		t.Broadcast(t.GetRunState())
	case StepInto:
		t.logger.Info("handle message", "case", "step presentation")
		if msg.targetStep < 0 || int(msg.targetStep) >= len(t.sections) {
			msg.targetStep = min(int32(len(t.sections)-1), max(int32(0), msg.targetStep))
		}

		t.step = msg.targetStep - 1
		t.timer = time.NewTimer(0)
	}

	return nil
}

func (t RunTask) GetRunState() RunStatusResponse {
	step := t.step
	if step < 0 {
		step = 0
	}
	if step >= int32(len(t.sections)) {
		step = int32(len(t.sections)) - 1
	}

	state := RunStatusResponse{
		State:  "running",
		Step:   t.sections[t.step],
		MsLeft: t.timerEnd.Sub(time.Now()).Milliseconds(),
	}

	if !t.isRunning {
		state.State = "stopped"
		state.MsLeft = t.timeRemaining.Milliseconds()
	}

	if t.timeRemaining == 0 && t.timerEnd.IsZero() {
		state.MsLeft = 0
	}

	return state
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
