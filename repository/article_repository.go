package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/SeladaKeju/article-service.git/domain"
)

type ArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(ctx context.Context, article domain.Article) (domain.Article, error) {
	const query = `
		INSERT INTO articles (id, author_id, title, body, created_at)
		SELECT $1, $2, $3, $4, $5
		WHERE EXISTS (SELECT 1 FROM authors WHERE id = $2)
		RETURNING id, author_id, title, body, created_at`

	var created domain.Article
	err := r.db.QueryRowContext(ctx, query, article.ID, article.AuthorID, article.Title, article.Body, article.CreatedAt).Scan(
		&created.ID, &created.AuthorID, &created.Title, &created.Body, &created.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Article{}, domain.ErrAuthorNotFound
	}
	if err == nil {
		created.CreatedAt = created.CreatedAt.UTC()
	}
	return created, err
}

type ListArticlesParams = domain.ListArticlesParams

func (r *ArticleRepository) List(ctx context.Context, params domain.ListArticlesParams) ([]domain.Article, error) {
	queryStr := `
		SELECT a.id, a.author_id, a.title, a.body, a.created_at
		FROM articles a
		JOIN authors au ON a.author_id = au.id`

	whereClauses := make([]string, 0, 3)
	args := make([]any, 0, 5)
	argIndex := 1

	if params.Author != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(au.name) = LOWER($%d)", argIndex))
		args = append(args, params.Author)
		argIndex++
	}

	if params.Query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(to_tsvector('simple', a.title || ' ' || a.body) @@ plainto_tsquery('simple', $%d) AND plainto_tsquery('simple', $%d) != ''::tsquery)", argIndex, argIndex))
		args = append(args, params.Query)
		argIndex++
	}

	if params.CursorTime != nil && params.CursorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(a.created_at, a.id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, *params.CursorTime, *params.CursorID)
		argIndex += 2
	}

	if len(whereClauses) > 0 {
		queryStr += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	queryStr += fmt.Sprintf(" ORDER BY a.created_at DESC, a.id DESC LIMIT $%d", argIndex)
	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]domain.Article, 0)
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(&a.ID, &a.AuthorID, &a.Title, &a.Body, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.CreatedAt = a.CreatedAt.UTC()
		articles = append(articles, a)
	}
	return articles, rows.Err()
}
