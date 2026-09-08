# Concurrency and Query Check

T07 check run locally on Windows with Docker Desktop and PostgreSQL 17 Alpine. The API uses one shared `*sql.DB`; request data is kept in controller/usecase local variables, while the repository holds only that safe-for-concurrent-use pool.

## Initial Configuration

These are implementation assumptions, not assessment requirements:

| Environment variable | Default | Purpose |
| --- | ---: | --- |
| `DB_MAX_OPEN_CONNS` | 10 | Maximum PostgreSQL connections in the API pool |
| `DB_MAX_IDLE_CONNS` | 5 | Connections retained idle in the pool |
| `REQUEST_TIMEOUT_SECONDS` | 5 | Deadline attached to each HTTP request context |

Values must be positive integers, and idle connections cannot exceed open connections. Tune them after measuring the deployment database and workload; they are not throughput targets.

The route timeout wraps `c.Request.Context()`, then controller, usecase, and repository pass that same context to `QueryContext`/`QueryRowContext`. A client cancellation or deadline therefore reaches PostgreSQL. No retry, lock, cache, or shared mutable request state is used.

## Automated Checks

Integration checks require `TEST_DATABASE_URL` pointing at a migrated test database. They create data with a unique marker and delete it during cleanup.

```powershell
$env:TEST_DATABASE_URL = 'postgres://USER:PASSWORD@localhost:5434/DATABASE?sslmode=disable'
go test -count=1 -v ./bootstrap ./repository -run 'Test(DatabaseHonorsQueryDeadline|DatabasePoolWaitHonorsDeadline|ArticleRepositoryConcurrentCreateAndList|ArticleRepositoryHonorsCanceledContext)'
```

Latest local run:

| Check | Result |
| --- | --- |
| Query deadline | `SELECT 1 FROM pg_sleep(1)` returned `context deadline exceeded`; subsequent `PingContext` succeeded |
| Pool wait | With a one-connection pool held by `db.Conn`, a second query timed out; after release, `PingContext` succeeded |
| Concurrent create/list | 20 creates + 20 list/search operations succeeded; all 20 response IDs were stored exactly once |
| Duration | 176.8113 ms total; 101.743717 ms average operation; 173.189 ms maximum operation; 226.2 ops/s over 40 operations |
| Errors | 0 |

The host Go toolchain is `windows/386`, so the race check ran in a Linux 64-bit Go container against the same separate test database. All packages passed with `go test -race -count=1 ./...`.

## Query Plans

`EXPLAIN (COSTS OFF)` was run after migrations against the local database:

| Query | Observed plan | Assessment |
| --- | --- | --- |
| Newest-first list | `Index Scan` on `idx_articles_created_at_id` | Order index used |
| Full-text search | `Bitmap Index Scan` on `idx_articles_search` | GIN search index used |
| Author filter | `Index Scan` on article order; `Seq Scan` on `authors` | Expected with only two seed authors; reassess `idx_authors_lower_name` on a larger author table |
| Author + search | `Bitmap Index Scan` on `idx_articles_author_created_at_id`, then full-text filter | Planner chose author-first access for this small dataset; reassess under representative production data |

This is a correctness and smoke-performance check, not a production benchmark. No k6 test or numerical SLA is claimed. Atomic create prevents partial rows but does not make a repeated client `POST` idempotent.
