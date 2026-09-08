package bootstrap

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDatabase(ctx context.Context, url string, maxOpenConns, maxIdleConns int) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
