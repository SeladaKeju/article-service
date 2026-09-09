package repository

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SeladaKeju/article-service.git/bootstrap"
	"github.com/SeladaKeju/article-service.git/domain"
)

func TestArticleRepositoryCreatePersistsArticle(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration test")
	}

	db, err := bootstrap.NewDatabase(context.Background(), url, 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	id, err := domain.NewUUIDv4()
	if err != nil {
		t.Fatal(err)
	}
	article := domain.Article{
		ID:        id,
		AuthorID:  "550e8400-e29b-41d4-a716-446655440000",
		Title:     "repository title",
		Body:      "repository body",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM articles WHERE id = $1", id)
	})

	created, err := NewArticleRepository(db).Create(context.Background(), article)
	if err != nil {
		t.Fatal(err)
	}
	if created != article {
		t.Fatalf("created = %#v, want %#v", created, article)
	}

	var count int
	if err := db.QueryRowContext(context.Background(), "SELECT count(*) FROM articles WHERE id = $1", id).Scan(&count); err != nil || count != 1 {
		t.Fatalf("persisted count = %d, err = %v", count, err)
	}

	article.ID, err = domain.NewUUIDv4()
	if err != nil {
		t.Fatal(err)
	}
	article.AuthorID = "3fa85f64-5717-4562-b3fc-2c963f66afa6"
	if _, err := NewArticleRepository(db).Create(context.Background(), article); !errors.Is(err, domain.ErrAuthorNotFound) {
		t.Fatalf("missing author error = %v", err)
	}
}

func TestArticleRepositoryListFilterAndPagination(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration test")
	}

	db, err := bootstrap.NewDatabase(context.Background(), url, 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewArticleRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	marker, err := domain.NewUUIDv4()
	if err != nil {
		t.Fatal(err)
	}
	marker = "t06" + strings.ReplaceAll(marker, "-", "")
	id1, _ := domain.NormalizeUUIDv4("00000000-0000-4000-8000-000000000001")
	id2, _ := domain.NormalizeUUIDv4("00000000-0000-4000-8000-000000000002")
	id3, _ := domain.NormalizeUUIDv4("00000000-0000-4000-8000-000000000003")

	a1 := domain.Article{ID: id1, AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: marker + " Go Concurrency", Body: "Concurrent requests in Go.", CreatedAt: now}
	a2 := domain.Article{ID: id2, AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: marker + " Go Patterns", Body: "Concurrency with independent requests.", CreatedAt: now}
	a3 := domain.Article{ID: id3, AuthorID: "6ba7b810-9dad-41d1-80b4-00c04fd430c8", Title: marker + " Golang Basics", Body: "Django comparison.", CreatedAt: now.Add(time.Second)}

	_, _ = db.ExecContext(ctx, "DELETE FROM articles WHERE id IN ($1, $2, $3)", id1, id2, id3)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM articles WHERE id IN ($1, $2, $3)", id1, id2, id3)
	})

	for _, a := range []domain.Article{a1, a2, a3} {
		if _, err := repo.Create(ctx, a); err != nil {
			t.Fatalf("failed to insert test article: %v", err)
		}
	}

	// Test whole-word search: 'go' matches a1 & a2, but not a3 (Golang/Django)
	res, err := repo.List(ctx, domain.ListArticlesParams{Query: marker + " go", Limit: 10})
	if err != nil || len(res) != 2 {
		t.Fatalf("query=go got %d results, err=%v; want 2", len(res), err)
	}

	// Test author case-insensitive filter
	res, err = repo.List(ctx, domain.ListArticlesParams{Query: marker, Author: "alice", Limit: 10})
	if err != nil || len(res) != 2 {
		t.Fatalf("author=alice got %d results, err=%v; want 2", len(res), err)
	}

	// Test combined AND filter
	res, err = repo.List(ctx, domain.ListArticlesParams{Query: marker + " go concurrency", Author: "ALICE", Limit: 10})
	if err != nil || len(res) != 2 {
		t.Fatalf("query=go concurrency & author=ALICE got %d results, err=%v", len(res), err)
	}

	// Test cursor pagination on same timestamp (id2 > id1, so id2 appears first)
	res, err = repo.List(ctx, domain.ListArticlesParams{Query: marker, Author: "alice", Limit: 1})
	if err != nil || len(res) != 1 || res[0].ID != id2 {
		t.Fatalf("page 1 got ID=%s, want %s", res[0].ID, id2)
	}

	res2, err := repo.List(ctx, domain.ListArticlesParams{Query: marker, Author: "alice", Limit: 1, CursorTime: &res[0].CreatedAt, CursorID: &res[0].ID})
	if err != nil || len(res2) != 1 || res2[0].ID != id1 {
		t.Fatalf("page 2 got ID=%s, want %s", res2[0].ID, id1)
	}

	// Empty result
	res, err = repo.List(ctx, domain.ListArticlesParams{Query: marker, Author: "Nobody", Limit: 10})
	if err != nil || len(res) != 0 {
		t.Fatalf("author=Nobody got %d results, want 0", len(res))
	}
}
