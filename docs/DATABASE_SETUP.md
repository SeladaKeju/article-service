# Database Setup

`docker compose up --build` waits for PostgreSQL, then runs every numbered `.up.sql` migration in `migrations/`. Migration `000003` seeds Alice and Bob idempotently.

For a new migration, add the next numbered `.up.sql` file and run:

```powershell
docker compose run --rm migrate
```

`docker compose down` keeps the named PostgreSQL volume. Use `docker compose down -v` only to deliberately reset all local database data; the next startup recreates the schema and seeds.

Integration tests must use a separate database via `TEST_DATABASE_URL`, never the application database.
