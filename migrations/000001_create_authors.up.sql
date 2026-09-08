CREATE TABLE authors (
    id TEXT PRIMARY KEY CHECK (id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'),
    name TEXT NOT NULL CHECK (btrim(name) <> '')
);
