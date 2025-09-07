package test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"contract-analysis-service/internal/pkg/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

type TestDBConfig struct {
	Dialect string
	Name   string
	LogMode bool
}

// GetDialect returns the database dialect
func (c TestDBConfig) GetDialect() string {
	if c.Dialect == "" {
		return "postgres"
	}
	return c.Dialect
}

// GetName returns the database name or DSN
func (c TestDBConfig) GetName() string {
	return c.Name
}

// GetLogMode returns whether to enable SQL logging
func (c TestDBConfig) GetLogMode() bool {
	return c.LogMode
}

func SetupTestDB(t *testing.T) *gorm.DB {
	ctx := context.Background()
	
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithStartupTimeout(30 * time.Second),
	}
	
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	
	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}
	
	host, err := pgContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	
	// Some environments return "localhost" which may resolve to ::1 first; use 127.0.0.1 to avoid IPv6 issues
	if host == "localhost" {
		host = "127.0.0.1"
	}
	
	// Create a DSN string for the database connection
	dsn := fmt.Sprintf("host=%s port=%d user=testuser password=testpass dbname=testdb sslmode=disable", 
		host, port.Int())
	
	cfg := TestDBConfig{
		Dialect: "postgres",
		Name:    dsn,
		LogMode: true,
	}
	
	db, err := database.NewDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	
	// Cleanup function
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})
	
	return db
}
