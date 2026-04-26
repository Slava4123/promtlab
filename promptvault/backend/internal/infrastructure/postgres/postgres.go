package postgres

import (
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/config"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.DatabaseConfig, isDev bool) (*gorm.DB, error) {
	logLevel := logger.Warn
	if isDev {
		logLevel = logger.Info
	}
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	// OpenTelemetry: span на каждый SQL query. No-op если TracerProvider — default.
	// Phase 16 Этап 3 — distributed tracing через GORM hooks.
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		slog.Warn("postgres.otel_plugin_failed", "error", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	maxOpen := cfg.MaxOpenConns
	if maxOpen == 0 {
		maxOpen = 25
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle == 0 {
		maxIdle = 5
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
