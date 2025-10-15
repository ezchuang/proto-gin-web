package platform

import "testing"

func TestLoadDefaults(t *testing.T) {
	// ensure empty values trigger fallbacks
	t.Setenv("APP_ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("BASE_URL", "")
	t.Setenv("ADMIN_USER", "")

	cfg := Load()

	if cfg.Env != "development" {
		t.Fatalf("expected default Env 'development', got %q", cfg.Env)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default Port '8080', got %q", cfg.Port)
	}
	if cfg.DBUser != "proto_user" {
		t.Fatalf("expected default DBUser 'proto_user', got %q", cfg.DBUser)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Fatalf("expected default BaseURL 'http://localhost:8080', got %q", cfg.BaseURL)
	}
	if cfg.AdminUser != "admin" {
		t.Fatalf("expected default AdminUser 'admin', got %q", cfg.AdminUser)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("PORT", "9090")
	t.Setenv("POSTGRES_USER", "tester")
	t.Setenv("BASE_URL", "https://example.com")
	t.Setenv("ADMIN_USER", "root")

	cfg := Load()

	if cfg.Env != "staging" {
		t.Fatalf("expected Env override 'staging', got %q", cfg.Env)
	}
	if cfg.Port != "9090" {
		t.Fatalf("expected Port override '9090', got %q", cfg.Port)
	}
	if cfg.DBUser != "tester" {
		t.Fatalf("expected DBUser override 'tester', got %q", cfg.DBUser)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Fatalf("expected BaseURL override 'https://example.com', got %q", cfg.BaseURL)
	}
	if cfg.AdminUser != "root" {
		t.Fatalf("expected AdminUser override 'root', got %q", cfg.AdminUser)
	}
}
