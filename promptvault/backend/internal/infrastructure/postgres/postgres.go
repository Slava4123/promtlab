package postgres

import (
	"context"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/config"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PGCapabilities — сводка опциональных расширений PostgreSQL, обнаруженных
// при старте сервера. Используется для feature-gating функций, требующих
// extensions, которые могут быть недоступны на managed PG (e.g. Timeweb).
type PGCapabilities struct {
	// TrgmAvailable — установлен ли pg_trgm. Нужен для PossibleDuplicates
	// в analytics.Service и для fuzzy match в search.
	TrgmAvailable bool
}

// DetectExtensions проверяет наличие опциональных PostgreSQL расширений.
// Безопасно вызывать после Connect и RunMigrations — если миграция не смогла
// создать extension (нет прав на managed PG), DetectExtensions просто вернёт
// TrgmAvailable=false, и зависимые фичи graceful-deграднут.
func DetectExtensions(ctx context.Context, db *gorm.DB) (PGCapabilities, error) {
	var trgmCount int64
	if err := db.WithContext(ctx).
		Raw("SELECT COUNT(*) FROM pg_extension WHERE extname = 'pg_trgm'").
		Scan(&trgmCount).Error; err != nil {
		return PGCapabilities{}, err
	}
	return PGCapabilities{
		TrgmAvailable: trgmCount > 0,
	}, nil
}

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
		maxOpen = 15 // CR-12: было 25, см. defaults в loader.go
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle == 0 {
		maxIdle = 5
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	// CR-12: SetConnMaxLifetime короче (30m вместо 1h) + добавлен
	// SetConnMaxIdleTime — освобождать idle connections, чтобы при пиковой
	// нагрузке (background loops + analytics + webhook'и одновременно)
	// pool не упирался в стенку.
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}
