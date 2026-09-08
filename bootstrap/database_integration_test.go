package bootstrap

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestDatabaseHonorsQueryDeadline(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration test")
	}

	db, err := NewDatabase(context.Background(), url, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	var ignored int
	err = db.QueryRowContext(ctx, "SELECT 1 FROM pg_sleep(1)").Scan(&ignored)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("query error = %v, want context deadline exceeded", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("PingContext() after cancellation = %v", err)
	}
}

func TestDatabasePoolWaitHonorsDeadline(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration test")
	}

	db, err := NewDatabase(context.Background(), url, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	var value int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&value)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("pool wait error = %v, want context deadline exceeded", err)
	}

	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	conn = nil
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("PingContext() after connection release = %v", err)
	}
}
