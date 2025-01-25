package server

import (
	"context"
	"log/slog"
	"sync"
	"time"

	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
)

type TasksState struct {
	logger               *slog.Logger
	queriesStore         *queries.Queries
	cleanSectionInterval time.Duration
}

func WithSectionOrderCleanInterval(d time.Duration) func(*TasksState) {
	return func(tc *TasksState) {
		tc.cleanSectionInterval = d.Abs()
	}
}

func RunTasks(
	ctx context.Context,
	logger *slog.Logger,
	queriesStore *queries.Queries,
	opts ...func(*TasksState),
) {
	conf := TasksState{
		logger:       logger,
		queriesStore: queriesStore,
	}

	for _, opt := range opts {
		opt(&conf)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		conf.RunCleanSectionOrder(ctx)
	}()

	wg.Wait()
}

func (c TasksState) RunCleanSectionOrder(ctx context.Context) {
	duration := c.cleanSectionInterval
	if duration.Microseconds() == 0 {
		duration = 1 * time.Minute
	}

	t := time.NewTicker(duration)
	for {
		select {
		case <-t.C:
			c.cleanSectionOrder()
		case <-ctx.Done():
			return
		}
	}
}

func (c TasksState) cleanSectionOrder() {
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.logger.Info("start clean sections order")

	err := c.queriesStore.CleanPositions(dbCtx)
	if err != nil {
		c.logger.Error("clean sections order", "err", err)
	}

	c.logger.Info("finish clean sections order")
}
