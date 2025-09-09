//go:build repo_integration
// +build repo_integration

package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"contract-analysis-service/internal/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TestDB struct {
	DB        *gorm.DB
	container testcontainers.Container
	logger    *zap.Logger
}

type Repositories struct {
	Contract       repository.ContractRepository
	Milestone      repository.MilestoneRepository
	RiskAssessment repository.RiskAssessmentRepository
}

// SetupTestDatabase starts a Postgres container and returns a TestDB with an open gorm.DB
func SetupTestDatabase(t *testing.T) *TestDB {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))
	t.Helper()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithStartupTimeout(60*time.Second),
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(10*time.Second),
		).WithDeadline(120 * time.Second),
	}

	pg, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := pg.Host(ctx)
	if err != nil {
		t.Fatalf("get host: %v", err)
	}

	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("get mapped port: %v", err)
	}

	// Use container IP instead of localhost to avoid Docker networking issues
	if host == "localhost" {
		host = "0.0.0.0"
	}

	dsn := fmt.Sprintf("host=%s port=%d user=testuser password=testpass dbname=testdb sslmode=disable", host, port.Int())
	
	// Add retry logic for database connection
	var db *gorm.DB
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			// Verify connection
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					break
				}
			}
		}
		
		if i == maxRetries-1 {
			_ = pg.Terminate(ctx)
			t.Fatalf("failed to connect to database after %d retries: %v", maxRetries, err)
		}
		
		time.Sleep(2 * time.Second)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		_ = pg.Terminate(ctx)
	})

	// Run migrations
	err = db.AutoMigrate(
		&repository.Contract{},
		&repository.Milestone{},
		&repository.RiskAssessment{},
	)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return &TestDB{
		DB:        db,
		container: pg,
		logger:    logger,
	}
}

// CreateRepositories constructs concrete Postgres-backed repositories on this DB
func (t *TestDB) CreateRepositories() Repositories {
	return Repositories{
		Contract:       repository.NewPostgresContractRepository(t.DB, t.logger),
		Milestone:      repository.NewPostgresMilestoneRepository(t.DB, t.logger),
		RiskAssessment: repository.NewPostgresRiskAssessmentRepository(t.DB, t.logger),
	}
}

// Cleanup truncates all tables to ensure test isolation
func (t *TestDB) Cleanup() {
	tables := []string{"risk_assessments", "milestones", "contracts"}
	for _, table := range tables {
		t.DB.Exec("TRUNCATE TABLE " + table + " CASCADE")
	}
}
