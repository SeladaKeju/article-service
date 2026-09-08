package bootstrap

import (
	"context"
	"database/sql"
)

type Application struct {
	Env *Env
	DB  *sql.DB
}

func App(ctx context.Context) (*Application, error) {
	env, err := NewEnv()
	if err != nil {
		return nil, err
	}

	db, err := NewDatabase(ctx, env.DatabaseURL, env.DBMaxOpenConns, env.DBMaxIdleConns)
	if err != nil {
		return nil, err
	}

	return &Application{Env: env, DB: db}, nil
}

func (app *Application) Close() error {
	return app.DB.Close()
}
