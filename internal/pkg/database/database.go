package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DatabaseConfig defines the interface for database configuration
type DatabaseConfig interface {
	GetDialect() string
	GetName() string
	GetLogMode() bool
}

// NewDB creates a new GORM DB instance with the specified dialect
func NewDB(cfg DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.GetDialect() {
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(cfg.GetName()), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
		}
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.GetName()), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database dialect: %s", cfg.GetDialect())
	}

	// Enable debug logging if configured
	if cfg.GetLogMode() {
		db = db.Debug()
	}

	return db, nil
}
