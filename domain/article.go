package domain

import (
	"context"
	"time"
)

// Article is an article stored by the service.
type Article struct {
	ID        string
	AuthorID  string
	Title     string
	Body      string
	CreatedAt time.Time
}

// ListArticlesParams contains repository filters and cursor pagination values.
type ListArticlesParams struct {
	Query      string
	Author     string
	Limit      int
	CursorTime *time.Time
	CursorID   *string
}

// ArticleRepository stores and lists articles.
type ArticleRepository interface {
	Create(context.Context, Article) (Article, error)
	List(context.Context, ListArticlesParams) ([]Article, error)
}
