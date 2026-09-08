INSERT INTO authors (id, name)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Alice'),
    ('6ba7b810-9dad-41d1-80b4-00c04fd430c8', 'Bob')
ON CONFLICT (id) DO NOTHING;
