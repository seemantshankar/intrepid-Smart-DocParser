package container

import (
	"testing"
	"context"
	"database/sql"
	"fmt"
	"time"

	"contract-analysis-service/configs"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestNewContainer(t *testing.T) {
	ctx := context.Background()

	pg, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		postgres.WithInitScripts(), // none
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() { _ = pg.Terminate(ctx) }()

	host, err := pg.Host(ctx)
	if err != nil { t.Fatalf("get host: %v", err) }
	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil { t.Fatalf("get mapped port: %v", err) }
	if host == "localhost" { host = "127.0.0.1" }

	// Wait until ready
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "testuser", "testpassword", host, port.Int(), "testdb")
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		db, openErr := sql.Open("postgres", dsn)
		if openErr == nil {
			if pingErr := db.Ping(); pingErr == nil { _ = db.Close(); break }
			_ = db.Close()
		}
		time.Sleep(300 * time.Millisecond)
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
		Jaeger: configs.JaegerConfig{URL: "", SamplingRate: 1.0},
		Logger: configs.LoggerConfig{Level: "debug"},
	}

	ctr := NewContainer(cfg)
	if ctr == nil {
		t.Fatal("expected non-nil container")
	}

	sqlDB, err := ctr.DB.DB()
	if err != nil { t.Fatalf("get sql db: %v", err) }
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil { t.Fatalf("db ping failed: %v", err) }
}
