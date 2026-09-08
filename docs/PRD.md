# Article Service PRD

Source: [Backend Engineer Assessment](<Backend Assessment.pdf>), pages 1-2. **Assumptions** are project defaults, not assessment requirements. Architecture, folders, and Docker usage follow the project owner's request.

## 1. Overview

Build a lightweight Go article service with keyword search and author filtering. The assessment allows three days and prioritizes quality and understanding of the solution.

## 2. Goals

- Fulfill all assessment requirements with simple, maintainable code.
- Keep search and filtering efficient as article volume grows.
- Handle concurrent requests reliably.
- Make future features easy to add without implementing them now.

## 3. Scope

Article creation, listing, search, author filtering, persistent storage, automated tests, and Docker-based local setup. Future features should fit the structure without being built now.

## 4. Functional Requirements

| Operation | Required behavior |
| --- | --- |
| `POST /articles` | Create and persist a new article. |
| `GET /articles` | List articles, newest first. |
| `GET /articles?query=...` | Search keywords across article title and body. |
| `GET /articles?author=...` | Filter articles by the author's name. |

**Assumptions - API details:**

The agreed T01 details and examples are defined in [API Contract](API_CONTRACT.md).

- REST/JSON: creation accepts exactly three nonblank strings: `author_id`, `title`, and `body`. Reject missing/null/wrong-type/extra fields and malformed JSON. Trim author IDs before validation; preserve title/body content. Authors are preloaded through seed SQL.
- IDs are lowercase UUID v4 strings stored as text. Generate article IDs and timestamps server-side; return UTC RFC 3339 timestamps with six fractional digits matching stored precision.
- Responses: `201` with `{data: article}`; `200` with `{data: [...], next_cursor: string|null}`, no total count. Empty results use `data: []`. Errors use `{error: {code, message}}`: `400` for `invalid_request`, `author_not_found`, `invalid_limit`, or `invalid_cursor`; `500 internal_error` without SQL details for internal/database failures.
- Search: case-insensitive whole-word token matching, requiring all terms across title/body. Full author names match case-insensitively. Combined filters apply AND; blank parameters act as omitted.
- Pagination uses optional `limit` (default 20, range 1-100) and cursor parameters. Trim query-parameter edges; empty values are omitted. Reject invalid limits/cursors. Order by `created_at DESC, id DESC`; encode the last returned timestamp/ID as unpadded Base64 URL-safe JSON. Fetch one extra result to determine whether `next_cursor` is needed; otherwise return null. Keep filters unchanged across pages and restart when they change. Pagination follows a live feed, not a cross-request snapshot.
- Nonempty search input that produces no tokens returns an empty list. Repeated POST requests are independent creates; idempotency is not promised.

## 5. Technical Requirements

- Use Go as the primary language.
- Use a widely used database/search engine; access the database through SQL queries, without an ORM.
- Include unit and/or integration tests.
- **Project requirement:** Provide a `Dockerfile` for the API and `compose.yaml` to run the API and PostgreSQL locally. Docker Compose is optional in the assessment but selected for this project by the owner.
- Disclose AI assistance; the candidate remains responsible for quality and understanding. This PRD, task breakdown, API contract, and existing health endpoint/test were AI-assisted.

**Assumption - stack:** Retain Gin; use PostgreSQL, a Go SQL driver, parameterized queries, and SQL migrations.

**Assumptions - Docker setup:** Use two services (`api`, `db`), a named volume for PostgreSQL data, and environment-based configuration with `.env.example`. Keep real credentials out of images and Git. Wait for the database healthcheck before starting the API using Compose's `service_healthy` dependency condition ([Docker reference](https://docs.docker.com/compose/how-tos/startup-order/)). Document configuration, SQL migration/author-seed steps, startup with `docker compose up --build`, and shutdown in the README.

## 6. Data Model

Preserve the schema shown in the assessment:

| Table | Fields |
| --- | --- |
| `articles` | `id` (text), `author_id` (text), `title` (text), `body` (text), `created_at` (timestamp) |
| `authors` | `id` (text), `name` (text) |

One author has many articles. IDs are primary keys; `articles.author_id` references `authors.id`. Author names are not assumed unique.

## 7. Architecture

Route → Controller → Usecase → Repository → Database

```text
cmd/             Application entry point
api/controller/  Decode requests, validate input, map responses/errors
api/route/       Register endpoints and connect controllers
bootstrap/      Load configuration, connect database, wire dependencies
domain/         Article/author models and necessary contracts
repository/     SQL persistence, search, filtering, and ordering
usecase/        Application rules and operation coordination
migrations/     SQL schema and index changes
```

Keep HTTP in controllers and SQL in repositories. Use direct dependency wiring and only necessary contracts.

## 8. Scalability & Concurrency Considerations

Efficient large-volume access, accurate search/filtering, and reliable concurrent operation are required. Dataset size, traffic targets, and latency thresholds are unspecified.

**Assumptions - implementation approach:**

- Filter, sort, and paginate in SQL. Use PostgreSQL full-text search over title/body with a GIN index and `simple` text configuration.
- Index `lower(authors.name)` for case-insensitive full-name filtering, article creation order, and `(author_id, created_at, id)`; validate choices against representative query plans.
- Use a shared connection pool, request cancellation/timeouts, database constraints, atomic inserts, and request-local state.
- Use cursor pagination to avoid deep offsets. Measure representative loads; record dataset size, concurrency, latency, and errors without inventing acceptance thresholds.

## 9. Testing Strategy

**Proposed checks:** Use Go's testing tools and HTTP test utilities for creation, newest-first ordering, title/body search, author and combined filters, empty results, validation, and pagination with tied timestamps.

Integration-test persistence, SQL search, and concurrent creates/reads: successful inserts persist once without partial records. Run the race detector where supported and a representative load/query-plan check. Retain the existing health smoke test.

Verify Docker setup from a fresh environment, including migrations and seeded authors, then check health/create/list endpoints. Confirm stored articles survive container recreation while retaining the database volume.

## 10. Non-Goals

Article update/delete/detail endpoints, author-management APIs, authentication, frontend UI, categories, comments, fuzzy search, and relevance ranking. Caches, separate search clusters, queues, microservices, and distributed locks are deferred until justified by requirements or measurements.

## 11. Acceptance Criteria

- Both required article endpoints work; saved articles are retrievable newest first.
- Keyword search covers title and body; author-name filtering returns matching authors' articles.
- Storage follows the supplied articles/authors relationship; the implementation uses Go and SQL without an ORM.
- Automated tests pass; proposed search/concurrency checks document workload, results, and limitations.
- Code follows the requested layers and global folders, with no speculative future-feature implementation.
- Assumptions are documented and tested if adopted.
- Following the README, a reviewer can build and run the API with PostgreSQL through Docker Compose, apply migrations/seeds, exercise the endpoints, and retain data across container recreation.
- Submission meets the three-day timeframe, discloses AI-assisted work, and can be explained by the candidate.
