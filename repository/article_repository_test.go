package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/SeladaKeju/article-service.git/bootstrap"
	"github.com/SeladaKeju/article-service.git/domain"
)

func TestArticleRepositoryCreatePersistsArticle(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is required for PostgreSQL integration test")
	}

	db, err := bootstrap.OpenDatabase(context.Background(), url)
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
	if _, err := NewArticleRepository(db).Create(context.Background(), article); !errors.Is(err, ErrAuthorNotFound) {
		t.Fatalf("missing author error = %v", err)
	}
}
