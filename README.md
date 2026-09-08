# Article Service

Go, Gin, and PostgreSQL service for creating, listing, searching, and filtering articles.

## Prerequisites and start

Install Docker Desktop (including Docker Compose), then create a local environment file and replace its placeholder values:

```powershell
Copy-Item .env.example .env
docker compose up --build
```

Compose starts PostgreSQL, waits until it is healthy, applies migrations and author seeds, then starts the API at `http://localhost:8080`. PostgreSQL is exposed at the `POSTGRES_HOST_PORT` value (the example uses `localhost:5434`); its port inside Docker is `5432`.

The values in `.env` are local credentials. Do not commit that file. Configuration is intentionally limited to:

| Variable | Purpose |
| --- | --- |
| `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` | Database name and local credentials |
| `POSTGRES_HOST_PORT` | PostgreSQL port exposed on the host |
| `API_PORT` | API port exposed on the host |
| `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS` | Shared database-pool limits (defaults: 10 and 5) |
| `REQUEST_TIMEOUT_SECONDS` | Per-request database deadline (default: 5) |

Seeded authors are available after startup:

| Name | ID |
| --- | --- |
| Alice | `550e8400-e29b-41d4-a716-446655440000` |
| Bob | `6ba7b810-9dad-41d1-80b4-00c04fd430c8` |

## Use the API

Create an article with a seeded author:

```powershell
Invoke-RestMethod http://localhost:8080/articles -Method Post -ContentType 'application/json' -Body '{"author_id":"550e8400-e29b-41d4-a716-446655440000","title":"Go Concurrency","body":"Concurrent requests in Go."}'
```

List, search, or combine filters:

```text
GET /articles
GET /articles?query=go%20concurrency&limit=20
GET /articles?author=alice
GET /articles?query=go&author=alice&limit=1&cursor=<next_cursor>
```

Keep the same filters with `next_cursor`; restart without a cursor when any filter changes. The full request, response, error, cursor, and pagination contract is in [API Contract](docs/API_CONTRACT.md).

## Database and architecture

Migrations live in `migrations/` and are applied automatically by Compose. For a later numbered `.up.sql` migration:

```powershell
docker compose run --rm migrate
```

`docker compose down` stops the stack while keeping database data. `docker compose down -v` is an intentional destructive local reset. See [Database Setup](docs/DATABASE_SETUP.md) for details.

Request flow is `Route -> Controller -> Usecase -> Repository -> PostgreSQL`. Controllers handle HTTP, use cases hold application rules, and repositories contain parameterized SQL. PostgreSQL full-text search uses `simple` configuration with a GIN index; lower-cased author-name and cursor-order indexes support the remaining filters. This keeps search, ordering, and pagination in the database without adding an ORM or separate search service.

## Tests and operational notes

Run unit tests and static checks:

```powershell
go test ./...
go vet ./...
```

Integration tests require a separate `TEST_DATABASE_URL`; setup, race-detector results, and Compose persistence verification are in [Final Verification](docs/FINAL_VERIFICATION.md). Connection-pool settings, concurrent create/list results, query-plan checks, and known limits are in [Concurrency and Query Check](docs/CONCURRENCY.md).

Pagination is a live feed rather than a cross-request snapshot, and repeated `POST` requests are not idempotent. k6, caching, queues, a search cluster, and distributed locks are intentionally out of scope until measurements require them.

## AI assistance

AI assisted the PRD, task breakdown, API documentation, implementation, and test/documentation drafts. The candidate remains responsible for reviewing, running, and explaining the code and its trade-offs.
