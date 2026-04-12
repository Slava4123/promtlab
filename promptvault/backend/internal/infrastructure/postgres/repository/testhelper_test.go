package repository

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"promptvault/internal/models"
)

// setupTestDB starts a real PostgreSQL container and returns a GORM *gorm.DB
// connected to it. The container is terminated via t.Cleanup.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := gorm.Open(pgdriver.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	// Run migrations for all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.TeamMember{},
		&models.Collection{},
		&models.Prompt{},
		&models.Tag{},
		&models.PromptVersion{},
		&models.EmailVerification{},
		&models.LinkedAccount{},
		&models.TeamInvitation{},
		&models.APIKey{},
		&models.UserBadge{},
		&models.UserTOTP{},
		&models.UserTOTPBackupCode{},
		&models.AuditLog{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}
