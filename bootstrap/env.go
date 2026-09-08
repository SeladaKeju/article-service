package bootstrap

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Env struct {
	DatabaseURL    string
	Address        string
	DBMaxOpenConns int
	DBMaxIdleConns int
	RequestTimeout time.Duration
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

	maxOpenConns, err := positiveIntEnv("DB_MAX_OPEN_CONNS", 10)
	if err != nil {
		return nil, err
	}
	maxIdleConns, err := positiveIntEnv("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	if maxIdleConns > maxOpenConns {
		return nil, fmt.Errorf("DB_MAX_IDLE_CONNS must not exceed DB_MAX_OPEN_CONNS")
	}
	timeoutSeconds, err := positiveIntEnv("REQUEST_TIMEOUT_SECONDS", 5)
	if err != nil {
		return nil, err
	}

	return &Env{
		DatabaseURL:    databaseURL,
		Address:        address,
		DBMaxOpenConns: maxOpenConns,
		DBMaxIdleConns: maxIdleConns,
		RequestTimeout: time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

func positiveIntEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}
