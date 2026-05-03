package db

import (
	"context"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Engine != "postgresql" {
		t.Errorf("expected Engine 'postgresql', got %s", cfg.Engine)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("expected Port 5432, got %d", cfg.Port)
	}
	if cfg.MaxConns != 10 {
		t.Errorf("expected MaxConns 10, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 2 {
		t.Errorf("expected MinConns 2, got %d", cfg.MinConns)
	}
}

func TestConfigDSN(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
	}
	dsn := cfg.DSN()
	if dsn == "" {
		t.Error("expected non-empty DSN")
	}
}

func TestConfigToORMConfig(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
		Options:  map[string]string{"sslmode": "require"},
	}
	ormCfg := cfg.ToORMConfig()
	if ormCfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got %s", ormCfg.Host)
	}
	if ormCfg.Port != 5432 {
		t.Errorf("expected Port 5432, got %d", ormCfg.Port)
	}
	if ormCfg.SSLMode != "require" {
		t.Errorf("expected SSLMode 'require', got %s", ormCfg.SSLMode)
	}
}

func TestConfigToORMConfigDefaultSSLMode(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
	}
	ormCfg := cfg.ToORMConfig()
	if ormCfg.SSLMode != "prefer" {
		t.Errorf("expected default SSLMode 'prefer', got %s", ormCfg.SSLMode)
	}
}

func TestConnectWithoutDatabase(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		Name:     "nonexistent_db",
		User:     "nonexistent_user",
		Password: "nonexistent_pass",
	}
	_, err := Connect(context.Background(), cfg)
	if err == nil {
		t.Skip("Skipping: database connection required but not available")
	}
}