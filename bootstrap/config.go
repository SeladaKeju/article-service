package bootstrap

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	Address     string
}

func LoadConfig() (Config, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	address := os.Getenv("ADDRESS")
	if address == "" {
		address = ":8080"
	}

	return Config{DatabaseURL: url, Address: address}, nil
}
