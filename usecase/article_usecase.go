package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/SeladaKeju/article-service.git/domain"
	"github.com/SeladaKeju/article-service.git/repository"
)

var ErrInvalidRequest = errors.New("invalid request")

type ArticleStore interface {
	Create(context.Context, domain.Article) (domain.Article, error)
}

type CreateArticleInput struct {
	AuthorID string
	Title    string
	Body     string
}

type ArticleUsecase struct {
	store ArticleStore
}

func NewArticleUsecase(store ArticleStore) *ArticleUsecase {
	return &ArticleUsecase{store: store}
}

func (u *ArticleUsecase) Create(ctx context.Context, input CreateArticleInput) (domain.Article, error) {
	authorID, valid := domain.NormalizeUUIDv4(strings.TrimSpace(input.AuthorID))
	if !valid || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Body) == "" {
		return domain.Article{}, ErrInvalidRequest
	}

	id, err := domain.NewUUIDv4()
	if err != nil {
		return domain.Article{}, err
	}

	return u.store.Create(ctx, domain.Article{
		ID:        id,
		AuthorID:  authorID,
		Title:     input.Title,
		Body:      input.Body,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	})
}

func IsAuthorNotFound(err error) bool {
	return errors.Is(err, repository.ErrAuthorNotFound)
}
