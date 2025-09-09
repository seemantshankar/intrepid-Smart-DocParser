package container

import (
	"context"
	"fmt"
	"database/sql"
	"testing"
	"time"

	"contract-analysis-service/configs"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	_ "github.com/lib/pq"
)

func TestNewContainer_WithPostgres(t *testing.T) {
	ctx := context.Background()

	pg, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		postgres.WithInitScripts(), // no scripts, just default
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		_ = pg.Terminate(ctx)
	}()

	host, err := pg.Host(ctx)
	if err != nil {
		t.Fatalf("get host: %v", err)
	}
	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("get mapped port: %v", err)
	}

	// Some environments return "localhost" which may resolve to ::1 first; use 127.0.0.1 to avoid IPv6 issues
	if host == "localhost" {
		host = "127.0.0.1"
	}

	// Wait for DB to be ready by pinging via database/sql
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		"testuser", "testpassword", host, port.Int(), "testdb")
	ready := false
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		db, openErr := sql.Open("postgres", dsn)
		if openErr == nil {
			pingErr := db.Ping()
			db.Close()
			if pingErr == nil {
				ready = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("postgres not ready after timeout at %s:%d", host, port.Int())
	}

	cfg := &configs.Config{
		Environment: "test",
		ServiceName: "contract-analysis-service",
		Server:      configs.ServerConfig{Port: "8080"},
		Database: configs.DatabaseConfig{
			Dialect: "postgres",
			Name:   fmt.Sprintf("host=%s port=%d user=testuser password=testpassword dbname=testdb sslmode=disable", host, port.Int()),
			LogMode: true,
		},
		Jaeger: configs.JaegerConfig{
			URL:          "", // no exporter in tests
			SamplingRate: 1.0,
		},
		Logger: configs.LoggerConfig{Level: "debug"},
	}

	ctr := NewContainer(cfg)
	if ctr == nil {
		t.Fatal("expected container, got nil")
	}

	// Basic ping to ensure DB is usable
	sqlDB, err := ctr.DB.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("db ping failed: %v", err)
	}

	fmt.Println("integration container initialized successfully")
}
