package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
)

func ListenAndServe(ctx context.Context, wg *sync.WaitGroup, addr string, logger *slog.Logger, queries *queries.Queries) {
	server := http.Server{
		Addr:         addr,
		Handler:      routes(logger, queries),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	wg.Add(1)
	go closeServer(ctx, wg, logger, &server)

	logger.InfoContext(ctx, "server listening", "on", addr)
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server exited unexpectedly", "err", err)
		return
	}
}

func closeServer(ctx context.Context, wg *sync.WaitGroup, logger *slog.Logger, server *http.Server) {
	defer wg.Done()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("error shutting down server", "err", err)
		return
	}
}
