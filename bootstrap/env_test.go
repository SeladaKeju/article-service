package bootstrap

import (
	"testing"
	"time"
)

func TestNewEnvDefaultsConcurrencySettings(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://article:article@localhost:5432/article_service")
	t.Setenv("ADDRESS", "")
	t.Setenv("DB_MAX_OPEN_CONNS", "")
	t.Setenv("DB_MAX_IDLE_CONNS", "")
	t.Setenv("REQUEST_TIMEOUT_SECONDS", "")

	env, err := NewEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env.DBMaxOpenConns != 10 || env.DBMaxIdleConns != 5 || env.RequestTimeout != 5*time.Second {
		t.Fatalf("unexpected concurrency settings: %#v", env)
	}
}

func TestNewEnvRejectsInvalidConcurrencySettings(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://article:article@localhost:5432/article_service")
	t.Setenv("DB_MAX_OPEN_CONNS", "2")
	t.Setenv("DB_MAX_IDLE_CONNS", "3")

	if _, err := NewEnv(); err == nil {
		t.Fatal("NewEnv() error = nil")
	}
}
