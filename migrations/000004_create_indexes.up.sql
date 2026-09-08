CREATE INDEX idx_articles_search ON articles USING gin (to_tsvector('simple', title || ' ' || body));
CREATE INDEX idx_authors_lower_name ON authors (lower(name));
CREATE INDEX idx_articles_created_at_id ON articles (created_at DESC, id DESC);
CREATE INDEX idx_articles_author_created_at_id ON articles (author_id, created_at DESC, id DESC);
