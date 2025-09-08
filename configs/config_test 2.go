package configs

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configContent := `
server:
  port: "8080"
database:
  host: "localhost"
logger:
  level: "info"
`
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Server.Port)
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("expected level info, got %s", cfg.Logger.Level)
	}
}
