package domain

import (
	"context"
	"time"
)

// Article is an article stored by the service.
type Article struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id" binding:"required"`
	Title     string    `json:"title" binding:"required"`
	Body      string    `json:"body" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
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

// ListArticlesInput contains client-provided article list parameters.
type ListArticlesInput struct {
	Query  string
	Author string
	Limit  string
	Cursor string
}

// ListArticlesResult contains one page of articles and an optional next cursor.
type ListArticlesResult struct {
	Articles   []Article
	NextCursor *string
}

// ArticleUsecase defines the article operations used by delivery layers.
type ArticleUsecase interface {
	Create(context.Context, Article) (Article, error)
	List(context.Context, ListArticlesInput) (ListArticlesResult, error)
}
