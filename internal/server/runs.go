package server

import (
	"context"
	"database/sql"
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

			switch string(p) {
			case "start":
				runs[ID].StartPresentation()
			}

		}
	})
}

type RunTask struct {
	presentationID int64
	logger         *slog.Logger
	queriesStore   *queries.Queries
	conns          map[string]*websocket.Conn
	// runs state
	ctx    context.Context
	cancel context.CancelFunc
	timer  *time.Timer
	step   int32
}

func NewRun(presentationID int64, logger *slog.Logger, queriesStore *queries.Queries) RunTask {
	return RunTask{
		presentationID: presentationID,
		logger:         logger,
		queriesStore:   queriesStore,
		conns:          make(map[string]*websocket.Conn),
	}
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

func (t RunTask) StartPresentation() {
	if t.ctx != nil {
		t.cancel()
	}

	t.step = 0
	t.timer = time.NewTimer(0)
	t.ctx, t.cancel = context.WithCancel(context.Background())

	t.logger.Info("runs start", "ID", t.presentationID)
	go t.Run()
}

func (t RunTask) Run() {
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-t.timer.C:
			t.step += 1

			dbCtx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
			section, err := t.queriesStore.GetSectionByPosition(
				dbCtx,
				queries.GetSectionByPositionParams{
					PresentationID: t.presentationID,
					Step:           t.step,
				},
			)
			cancel()

			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
					return
				default:
					return
				}
			}

			t.timer = time.NewTimer(section.Duration)
			t.Broadcast(map[string]any{
				"step":     section.ID,
				"duration": section.Duration.String(),
			})
			t.logger.Info("step", "section", section)
		}
	}
}

func (t RunTask) Broadcast(msg any) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pm, err := websocket.NewPreparedMessage(websocket.BinaryMessage, b)
	if err != nil {
		return err
	}

	g := new(errgroup.Group)
	for _, conn := range t.conns {
		g.Go(func() error {
			return conn.WritePreparedMessage(pm)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
