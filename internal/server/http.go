package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func ListenAndServe(ctx context.Context, addr string, logger *slog.Logger) {
	server := http.Server{
		Addr:         addr,
		Handler:      routes(logger),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	logger.InfoContext(ctx, "server listening", "on", addr)
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server exited unexpectedly", "err", err)
		return
	}
}

func closeServer(ctx context.Context, logger *slog.Logger, server *http.Server) {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("error shutting down server", "err", err)
		return
	}
}
