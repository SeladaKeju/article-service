package bootstrap

import (
	"context"
	"database/sql"
)

// Application owns initialized runtime dependencies.
type Application struct {
	Env *Env
	DB  *sql.DB
}

// App initializes the environment and database connection.
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

// Close releases application resources.
func (app *Application) Close() error {
	return app.DB.Close()
}
