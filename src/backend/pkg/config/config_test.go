package config

import "testing"

func TestValidateConfigDefaultsPersistenceWorkers(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			DSN: "sqlite.db",
		},
		Session: SessionConfig{
			Secret: "test-secret",
		},
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}

	if cfg.Detection.PersistenceWorkers != 2 {
		t.Fatalf("expected default persistence_workers=2, got %d", cfg.Detection.PersistenceWorkers)
	}
}

func TestValidateConfigKeepsExplicitPersistenceWorkers(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			DSN: "sqlite.db",
		},
		Session: SessionConfig{
			Secret: "test-secret",
		},
		Detection: DetectionConfig{
			PersistenceWorkers: 4,
		},
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}

	if cfg.Detection.PersistenceWorkers != 4 {
		t.Fatalf("expected persistence_workers to stay 4, got %d", cfg.Detection.PersistenceWorkers)
	}
}
