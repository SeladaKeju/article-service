CREATE TABLE articles (
    id TEXT PRIMARY KEY CHECK (id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'),
    author_id TEXT NOT NULL REFERENCES authors (id),
    title TEXT NOT NULL CHECK (btrim(title) <> ''),
    body TEXT NOT NULL CHECK (btrim(body) <> ''),
    created_at TIMESTAMPTZ NOT NULL
);
