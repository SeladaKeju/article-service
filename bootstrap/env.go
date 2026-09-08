package bootstrap

import (
	"fmt"
	"os"
)

type Env struct {
	DatabaseURL string
	Address     string
}

func NewEnv() (*Env, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	address := os.Getenv("ADDRESS")
	if address == "" {
		address = ":8080"
	}

	return &Env{DatabaseURL: databaseURL, Address: address}, nil
}
