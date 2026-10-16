package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"crmdemo/internal/analytics"
	"crmdemo/internal/blogger"
	"crmdemo/internal/campaign"
	"crmdemo/internal/config"
	"crmdemo/internal/db"
	"crmdemo/internal/httpx"
	"crmdemo/internal/ozon"
	"crmdemo/internal/seed"
	"crmdemo/internal/user"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("svc", "api")
	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL, log)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool, log); err != nil {
		return err
	}
	if cfg.DemoReset {
		if err := db.WithLock(ctx, pool, func(ctx context.Context) error {
			st, err := seed.Run(ctx, pool, time.Now())
			if err == nil {
				log.Info("demo data seeded", "bloggers", st.Bloggers, "campaigns", st.Campaigns,
					"integrations", st.Integrations, "metric_rows", st.MetricRows,
					"ozon_reports", st.OzonReports, "users", st.Users)
			}
			return err
		}); err != nil {
			return err
		}
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           newRouter(pool, log, cfg.CORSOrigin),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func newRouter(pool *pgxpool.Pool, log *slog.Logger, corsOrigin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpx.Error(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	blogger.NewHandler(blogger.NewRepository(pool), log).Register(mux)
	campaign.NewHandler(campaign.NewRepository(pool), log).Register(mux)
	analytics.NewHandler(analytics.NewRepository(pool), log).Register(mux)
	ozon.NewHandler(ozon.NewRepository(pool), log).Register(mux)
	user.NewHandler(pool, log).Register(mux)

	return httpx.Chain(mux, httpx.Recover(log), httpx.Logger(log), httpx.CORS(corsOrigin))
}
