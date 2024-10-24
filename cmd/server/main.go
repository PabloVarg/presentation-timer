package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
	"github.com/PabloVarg/presentation-timer/internal/server"
	"github.com/jackc/pgx/v5"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	run(logger)
}

func run(logger *slog.Logger) {
	var wg sync.WaitGroup

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	port, ok := os.LookupEnv("PORT")
	if !ok {
		logger.Error("no port configured", "env", "PORT")
		return
	}

	queriesStore, conn, err := createQueries(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to DB", "err", err)
		return
	}
	wg.Add(1)
	defer func() {
		defer wg.Done()

		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn.Close(closeCtx)
	}()

	server.ListenAndServe(ctx, &wg, fmt.Sprintf(":%s", port), logger, queriesStore)

	logger.Info("closing resources")
	wg.Wait()
}

func createQueries(ctx context.Context) (*queries.Queries, *pgx.Conn, error) {
	conn, err := pgx.Connect(
		ctx,
		fmt.Sprintf(
			"dbname=%s user=%s password=%s host=%s sslmode=%s",
			os.Getenv("POSTGRES_DB"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_SSLMODE"),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	return queries.New(conn), nil, nil
}
