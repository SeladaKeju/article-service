package domain

import (
	"context"
	"time"
)

type Author struct {
	ID   string
	Name string
}

type Article struct {
	ID        string
	AuthorID  string
	Title     string
	Body      string
	CreatedAt time.Time
}

type ListArticlesParams struct {
	Query      string
	Author     string
	Limit      int
	CursorTime *time.Time
	CursorID   *string
}

type ArticleRepository interface {
	Create(context.Context, Article) (Article, error)
	List(context.Context, ListArticlesParams) ([]Article, error)
}
