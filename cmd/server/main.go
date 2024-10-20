package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/PabloVarg/presentation-timer/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	run(logger)
}

func run(logger *slog.Logger) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	port, ok := os.LookupEnv("PORT")
	if !ok {
		logger.Error("no port configured", "env", "PORT")
		return
	}

	server.ListenAndServe(ctx, fmt.Sprintf(":%s", port), logger)
}
