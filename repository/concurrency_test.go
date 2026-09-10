package repository

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/SeladaKeju/article-service.git/bootstrap"
	"github.com/SeladaKeju/article-service.git/domain"
	"github.com/SeladaKeju/article-service.git/usecase"
)

func TestArticleRepositoryConcurrentCreateAndList(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration test")
	}

	db, err := bootstrap.NewDatabase(context.Background(), url, 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	marker, err := domain.NewUUIDv4()
	if err != nil {
		t.Fatal(err)
	}
	marker = "t07-" + marker
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM articles WHERE title LIKE $1", marker+"%")
	})

	const creates = 20
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	articleUsecase := usecase.NewArticleUsecase(NewArticleRepository(db))
	created := make(chan domain.Article, creates)
	errs := make(chan error, creates*2)
	durations := make(chan time.Duration, creates*2)
	var group sync.WaitGroup
	started := time.Now()

	for i := 0; i < creates; i++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			operationStarted := time.Now()
			article, err := articleUsecase.Create(ctx, domain.Article{
				AuthorID: "550e8400-e29b-41d4-a716-446655440000",
				Title:    fmt.Sprintf("%s concurrency %d", marker, index),
				Body:     "concurrent create verification",
			})
			if err != nil {
				errs <- err
				return
			}
			created <- article
			durations <- time.Since(operationStarted)
		}(i)

		group.Add(1)
		go func() {
			defer group.Done()
			operationStarted := time.Now()
			_, err := articleUsecase.List(ctx, domain.ListArticlesInput{Query: "concurrency", Limit: "100"})
			if err != nil {
				errs <- err
				return
			}
			durations <- time.Since(operationStarted)
		}()
	}

	group.Wait()
	close(created)
	close(errs)
	close(durations)
	for err := range errs {
		t.Fatal(err)
	}

	ids := make(map[string]struct{}, creates)
	for article := range created {
		ids[article.ID] = struct{}{}
	}
	if len(ids) != creates {
		t.Fatalf("successful create IDs = %d, want %d", len(ids), creates)
	}

	var total, max time.Duration
	operations := 0
	for duration := range durations {
		total += duration
		operations++
		if duration > max {
			max = duration
		}
	}
	elapsed := time.Since(started)
	t.Logf("operations=%d total=%s average=%s max=%s throughput=%.1f ops/s", operations, elapsed, total/time.Duration(operations), max, float64(operations)/elapsed.Seconds())

	for id := range ids {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT count(*) FROM articles WHERE id = $1", id).Scan(&count); err != nil || count != 1 {
			t.Fatalf("article %s count = %d, err = %v", id, count, err)
		}
	}
}

func TestArticleRepositoryHonorsCanceledContext(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration test")
	}

	db, err := bootstrap.NewDatabase(context.Background(), url, 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewArticleRepository(db).List(ctx, domain.ListArticlesParams{Limit: 1}); err == nil {
		t.Fatal("List() error = nil with canceled context")
	}
}
