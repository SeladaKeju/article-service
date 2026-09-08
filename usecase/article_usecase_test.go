package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/SeladaKeju/article-service.git/domain"
	"github.com/SeladaKeju/article-service.git/repository"
)

type mockStore struct {
	created domain.Article
	err     error
}

func (m *mockStore) Create(_ context.Context, a domain.Article) (domain.Article, error) {
	m.created = a
	if m.err != nil {
		return domain.Article{}, m.err
	}
	return a, nil
}

func TestArticleUsecaseCreateSuccess(t *testing.T) {
	store := &mockStore{}
	uc := NewArticleUsecase(store)

	input := CreateArticleInput{
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
		input CreateArticleInput
	}{
		{"invalid author_id", CreateArticleInput{AuthorID: "not-a-uuid", Title: "Title", Body: "Body"}},
		{"blank author_id", CreateArticleInput{AuthorID: "   ", Title: "Title", Body: "Body"}},
		{"blank title", CreateArticleInput{AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: "   ", Body: "Body"}},
		{"blank body", CreateArticleInput{AuthorID: "550e8400-e29b-41d4-a716-446655440000", Title: "Title", Body: "   "}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Create(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("got err = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func TestArticleUsecaseCreateAuthorNotFound(t *testing.T) {
	store := &mockStore{err: repository.ErrAuthorNotFound}
	uc := NewArticleUsecase(store)

	input := CreateArticleInput{
		AuthorID: "550e8400-e29b-41d4-a716-446655440000",
		Title:    "Title",
		Body:     "Body",
	}

	_, err := uc.Create(context.Background(), input)
	if !IsAuthorNotFound(err) {
		t.Fatalf("got err = %v, want author not found", err)
	}
}
