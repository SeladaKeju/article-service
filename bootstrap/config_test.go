package bootstrap

import "testing"

func TestLoadConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://article:article@localhost:5432/articles")
	t.Setenv("ADDRESS", "")

	config, err := LoadConfig()
	if err != nil || config.Address != ":8080" || config.DatabaseURL == "" {
		t.Fatalf("LoadConfig() = %#v, %v", config, err)
	}
}

func TestLoadConfigRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() error = nil")
	}
}
