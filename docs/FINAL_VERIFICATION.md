# Final Verification

## Separate Test Database

Integration tests use `TEST_DATABASE_URL`, never `DATABASE_URL`. The local verification database is `article_service_test`; migrations 1–4 and author seeds were applied before the test run.

```powershell
$env:TEST_DATABASE_URL = 'postgres://USER:PASSWORD@localhost:5434/article_service_test?sslmode=disable'
go test -count=1 ./...
go vet ./...
go build ./cmd
```

The run passed unit tests, PostgreSQL integration tests, concurrent create/list checks, deadline checks, create persistence, search/filtering, cursor pagination, and malformed create input checks.

## Clean Compose Check

A separate Compose project was started without using the application stack's volume:

```powershell
$env:API_PORT = '18080'
$env:POSTGRES_HOST_PORT = '15434'
docker compose -p article-service-t08 up --build -d
```

It created `article-service-t08_postgres_data`, applied migrations 1–4 and author seeds, and started the API. An article was created through `POST /articles`, found through `GET /articles?query=t08persistence&author=Alice`, then found again after container recreation. The persisted article ID was `dfc8b7b0-6dfd-4de2-98dc-8f09b3cff4ec`.

The test did not use `down -v`; the named volume remained intact.

## Race Detector

The host Go toolchain reports `-race is not supported on windows/386`, so the race check ran in a Linux 64-bit Go container against `article_service_test`:

```powershell
docker run --rm -e TEST_DATABASE_URL=... golang:1.25-alpine go test -race -count=1 ./...
```

All packages passed. The first Docker retry exposed a full local disk; only rebuildable Go/npm caches were cleared, then Docker Desktop was restarted. No project source or PostgreSQL volume was removed.
