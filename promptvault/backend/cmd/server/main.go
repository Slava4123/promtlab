package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/riandyrn/otelchi"

	"promptvault/internal/app"
	"promptvault/internal/delivery/http/utils"
	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/postgres"
	"promptvault/internal/infrastructure/telemetry"
	corsmw "promptvault/internal/middleware/cors"
	loggermw "promptvault/internal/middleware/logger"
	metricsmw "promptvault/internal/middleware/metrics"
	sentrymw "promptvault/internal/middleware/sentry"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	var handler slog.Handler
	if cfg.Server.IsProd() {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(handler))

	// Sentry init — только если явно включён. fail-open: если GlitchTip
	// недоступен, SDK логирует ошибку и продолжает работу в no-op режиме.
	if cfg.Sentry.Enabled {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.Sentry.Dsn,
			Environment:      cfg.Sentry.Environment,
			Release:          cfg.Sentry.Release,
			AttachStacktrace: true,
			TracesSampleRate: cfg.Sentry.TracesSampleRate,
			Debug:            cfg.Sentry.Debug,
			// BeforeSend — скраббинг PII: удаляем Authorization header и cookies
			// из events, чтобы JWT токены и session cookies не утекали в GlitchTip.
			BeforeSend: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
				if event.Request != nil {
					if event.Request.Headers != nil {
						delete(event.Request.Headers, "Authorization")
						delete(event.Request.Headers, "Cookie")
					}
					event.Request.Cookies = ""
				}
				return event
			},
		}); err != nil {
			slog.Error("sentry.init failed", "error", err)
			// Не падаем — backend должен работать даже без error tracking.
		} else {
			slog.Info("sentry.init", "environment", cfg.Sentry.Environment, "release", cfg.Sentry.Release, "traces_sample_rate", cfg.Sentry.TracesSampleRate)
		}
	}

	db, err := postgres.Connect(cfg.Database, cfg.Server.IsDev())
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	if err := postgres.RunMigrations(cfg.Database.DSN()); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	// OpenTelemetry tracing init (Phase 16 Этап 3). No-op если cfg.Telemetry.Enabled=false.
	telemetryShutdown, err := telemetry.Setup(context.Background(), cfg.Telemetry,
		cfg.Server.Environment, cfg.Sentry.Release)
	if err != nil {
		slog.Error("telemetry init failed", "error", err)
		os.Exit(1)
	}

	application := app.New(cfg, db)

	r := chi.NewRouter()
	// Sentry middleware ПЕРВЫМ — ловит panics через Repanic:true и прокидывает
	// их дальше в chimw.Recoverer, который возвращает 500. Порядок:
	//   sentry (capture) → logger → Recoverer (respond 500) → RequestID → CORS
	// No-op если sentry.Init не был вызван.
	if cfg.Sentry.Enabled {
		r.Use(sentrymw.Handler())
	}
	r.Use(loggermw.Middleware)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	// otelchi: auto-инструментация HTTP requests — span на каждый handler
	// с trace_id propagated в context. No-op если TracerProvider — default
	// no-op (Telemetry.Enabled=false).
	r.Use(otelchi.Middleware("promptvault-api", otelchi.WithChiRoutes(r)))
	r.Use(corsmw.Middleware(cfg))
	// HTTP metrics middleware (Phase 16+ observability) — РОВНО после CORS
	// и до MountRoutes, чтобы покрывать все handlers (включая /metrics само).
	// Path label нормализуется до Chi route pattern (см. metrics.routePattern).
	r.Use(metricsmw.Middleware)

	if cfg.Server.IsDev() {
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)
		slog.Info("pprof enabled at /debug/pprof/")
	}

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteOK(w, map[string]string{"status": "ok"})
	})

	application.MountRoutes(r)
	application.StartBackground()

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 5 * time.Minute, // увеличен для SSE-стриминга AI
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "port", cfg.Server.Port, "env", cfg.Server.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	application.Shutdown(15 * time.Second)

	// OpenTelemetry: drain pending spans перед exit. No-op если SDK не активен.
	if err := telemetryShutdown(shutdownCtx); err != nil {
		slog.Warn("telemetry shutdown error", "error", err)
	}

	// Flush отправляет все pending Sentry events перед exit. Если Sentry не
	// был инициализирован — no-op.
	if cfg.Sentry.Enabled {
		sentry.Flush(2 * time.Second)
	}

	slog.Info("server stopped")
}
