package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/SeladaKeju/article-service.git/domain"
	"github.com/SeladaKeju/article-service.git/repository"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidLimit   = errors.New("invalid limit")
	ErrInvalidCursor  = errors.New("invalid cursor")
)

type CreateArticleInput struct {
	AuthorID string
	Title    string
	Body     string
}

type ArticleUsecase struct {
	store domain.ArticleRepository
}

func NewArticleUsecase(store domain.ArticleRepository) *ArticleUsecase {
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

type ListArticlesInput struct {
	Query  string
	Author string
	Limit  string
	Cursor string
}

type ListArticlesResult struct {
	Articles   []domain.Article
	NextCursor *string
}

type cursorPayload struct {
	CreatedAt *string `json:"created_at"`
	ID        *string `json:"id"`
}

func (u *ArticleUsecase) List(ctx context.Context, input ListArticlesInput) (ListArticlesResult, error) {
	query := strings.TrimSpace(input.Query)
	author := strings.TrimSpace(input.Author)

	limit := 20
	trimmedLimit := strings.TrimSpace(input.Limit)
	if trimmedLimit != "" {
		val, err := strconv.Atoi(trimmedLimit)
		if err != nil || val < 1 || val > 100 {
			return ListArticlesResult{}, ErrInvalidLimit
		}
		limit = val
	}

	var cursorTime *time.Time
	var cursorID *string
	trimmedCursor := strings.TrimSpace(input.Cursor)
	if trimmedCursor != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(trimmedCursor)
		if err != nil {
			return ListArticlesResult{}, ErrInvalidCursor
		}
		decoder := json.NewDecoder(bytes.NewReader(decoded))
		decoder.DisallowUnknownFields()
		var payload cursorPayload
		if err := decoder.Decode(&payload); err != nil {
			return ListArticlesResult{}, ErrInvalidCursor
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			return ListArticlesResult{}, ErrInvalidCursor
		}
		if payload.CreatedAt == nil || payload.ID == nil {
			return ListArticlesResult{}, ErrInvalidCursor
		}
		t, err := time.Parse("2006-01-02T15:04:05.000000Z", *payload.CreatedAt)
		if err != nil || t.Format("2006-01-02T15:04:05.000000Z") != *payload.CreatedAt {
			return ListArticlesResult{}, ErrInvalidCursor
		}
		normID, valid := domain.NormalizeUUIDv4(*payload.ID)
		if !valid {
			return ListArticlesResult{}, ErrInvalidCursor
		}
		cursorTime = &t
		cursorID = &normID
	}

	articles, err := u.store.List(ctx, domain.ListArticlesParams{
		Query:      query,
		Author:     author,
		Limit:      limit + 1,
		CursorTime: cursorTime,
		CursorID:   cursorID,
	})
	if err != nil {
		return ListArticlesResult{}, err
	}

	var nextCursor *string
	if len(articles) > limit {
		lastArticle := articles[limit-1]
		articles = articles[:limit]
		payloadBytes, _ := json.Marshal(cursorPayload{
			CreatedAt: stringPtr(lastArticle.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000Z")),
			ID:        stringPtr(lastArticle.ID),
		})
		enc := base64.RawURLEncoding.EncodeToString(payloadBytes)
		nextCursor = &enc
	}

	if articles == nil {
		articles = []domain.Article{}
	}

	return ListArticlesResult{
		Articles:   articles,
		NextCursor: nextCursor,
	}, nil
}

func stringPtr(s string) *string {
	return &s
}
