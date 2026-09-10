package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/SeladaKeju/article-service.git/domain"
)

type mockStore struct {
	created    domain.Article
	err        error
	listParams domain.ListArticlesParams
	listReturn []domain.Article
}

func (m *mockStore) Create(_ context.Context, a domain.Article) (domain.Article, error) {
	m.created = a
	if m.err != nil {
		return domain.Article{}, m.err
	}
	return a, nil
}

func (m *mockStore) List(_ context.Context, params domain.ListArticlesParams) ([]domain.Article, error) {
	m.listParams = params
	if m.err != nil {
		return nil, m.err
	}
	return m.listReturn, nil
}

func TestArticleUsecaseCreateSuccess(t *testing.T) {
	store := &mockStore{}
	uc := NewArticleUsecase(store)

	input := domain.Article{
		AuthorID: " 550E8400-E29B-41D4-A716-446655440000 ",
		Title:    " Test Title ",
		Body:     " Test Body ",
	}

	result, err := uc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, valid := domain.NormalizeUUIDv4(result.ID); !valid {
		t.Fatalf("invalid generated ID: %s", result.ID)
	}
	if result.AuthorID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("author_id = %s, want normalized lower UUID", result.AuthorID)
	}
	if result.Title != " Test Title " || result.Body != " Test Body " {
		t.Errorf("title/body mismatch: %v", result)
	}
}

func TestArticleUsecaseCreateValidationErrors(t *testing.T) {
	uc := NewArticleUsecase(&mockStore{})

	tests := []struct {
		name  string
		input domain.Article
	}{
		{"invalid author_id", domain.Article{AuthorID: "not-a-uuid", Title: "Title", Body: "Body"}},
		{"blank author_id", domain.Article{AuthorID: "   ", Title: "Title", Body: "Body"}},
		{"blank title", domain.Article{AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: "   ", Body: "Body"}},
		{"blank body", domain.Article{AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: "Title", Body: "   "}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Create(context.Background(), tt.input)
			if !errors.Is(err, domain.ErrInvalidArticle) {
				t.Fatalf("got err = %v, want ErrInvalidArticle", err)
			}
		})
	}
}

func TestArticleUsecaseCreateAuthorNotFound(t *testing.T) {
	store := &mockStore{err: domain.ErrAuthorNotFound}
	uc := NewArticleUsecase(store)

	input := domain.Article{
		AuthorID: "550e8400-e29b-41d4-a716-446655440000",
		Title:    "Title",
		Body:     "Body",
	}

	_, err := uc.Create(context.Background(), input)
	if !errors.Is(err, domain.ErrAuthorNotFound) {
		t.Fatalf("got err = %v, want ErrAuthorNotFound", err)
	}
}

func TestArticleUsecaseListValidation(t *testing.T) {
	uc := NewArticleUsecase(&mockStore{})

	invalidLimits := []string{"0", "-1", "101", "1.5", "abc"}
	for _, l := range invalidLimits {
		t.Run("limit_"+l, func(t *testing.T) {
			_, err := uc.List(context.Background(), domain.ListArticlesInput{Limit: l})
			if !errors.Is(err, domain.ErrInvalidLimit) {
				t.Fatalf("got %v, want ErrInvalidLimit for limit %q", err, l)
			}
		})
	}

	invalidCursors := []string{
		"invalid-base64!!!",
		base64.RawURLEncoding.EncodeToString([]byte(`{"created_at":"2026-09-08T12:00:00Z"}`)),         // missing id
		base64.RawURLEncoding.EncodeToString([]byte(`{"id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}`)), // missing created_at
		base64.RawURLEncoding.EncodeToString([]byte(`{"created_at":"2026-09-08T12:00:00.123456Z","id":"not-uuid"}`)),
		base64.RawURLEncoding.EncodeToString([]byte(`{"created_at":"2026-09-08T12:00:00.123Z","id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}`)), // wrong subsecond precision
		base64.RawURLEncoding.EncodeToString([]byte(`{"created_at":"2026-09-08T12:00:00.123456Z","id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","extra":1}`)),
	}
	for _, c := range invalidCursors {
		t.Run("cursor", func(t *testing.T) {
			_, err := uc.List(context.Background(), domain.ListArticlesInput{Cursor: c})
			if !errors.Is(err, domain.ErrInvalidCursor) {
				t.Fatalf("got %v, want ErrInvalidCursor for cursor %q", err, c)
			}
		})
	}
}

func TestArticleUsecaseListPagination(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	store := &mockStore{
		listReturn: []domain.Article{
			{ID: "3fa85f64-5717-4562-b3fc-2c963f66afa6", AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: "A1", Body: "B1", CreatedAt: now},
			{ID: "2fa85f64-5717-4562-b3fc-2c963f66afa6", AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: "A2", Body: "B2", CreatedAt: now},
		},
	}
	uc := NewArticleUsecase(store)

	res, err := uc.List(context.Background(), domain.ListArticlesInput{Limit: "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Articles) != 1 {
		t.Fatalf("articles count = %d, want 1", len(res.Articles))
	}
	if res.NextCursor == nil {
		t.Fatal("expected next_cursor, got nil")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(*res.NextCursor)
	if err != nil {
		t.Fatalf("cursor decode err: %v", err)
	}
	expectedPayload := `{"created_at":"` + now.Format("2006-01-02T15:04:05.000000Z") + `","id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}`
	if string(decoded) != expectedPayload {
		t.Fatalf("decoded cursor = %s, want %s", string(decoded), expectedPayload)
	}
}
