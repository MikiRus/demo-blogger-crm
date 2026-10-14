package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"crmdemo/internal/config"
	"crmdemo/internal/db"
	"crmdemo/internal/worker"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("svc", "worker")
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL, log)
	if err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Миграции идемпотентны и под advisory-локом, поэтому воркер
	// может стартовать раньше API и не упасть на пустой базе.
	if err := db.Migrate(ctx, pool, log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
	worker.New(pool, log, cfg.WorkerEvery).Run(ctx)
}
