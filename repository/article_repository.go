package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/SeladaKeju/article-service.git/domain"
)

var ErrAuthorNotFound = errors.New("author not found")

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
		return domain.Article{}, ErrAuthorNotFound
	}
	return created, err
}
