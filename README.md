# Article Service

Go article service backed by PostgreSQL.

## Run with Docker

Copy the example environment file, then start the stack:

```powershell
Copy-Item .env.example .env
docker compose up --build
```

Compose starts PostgreSQL, waits for its healthcheck, runs SQL migrations and author seeds, then starts the API at `http://localhost:8080`.

PostgreSQL is available from the host at `localhost:5434` (`article_service` / `article`); its container port remains `5432`.

Stop containers with `docker compose down`. This preserves PostgreSQL data. Use `docker compose down -v` only when intentionally resetting local data.

## Apply Later Migrations

Create a new numbered `.up.sql` file in `migrations/`, then run:

```powershell
docker compose run --rm migrate
```

See [database setup](docs/DATABASE_SETUP.md) for seed IDs and reset instructions.
