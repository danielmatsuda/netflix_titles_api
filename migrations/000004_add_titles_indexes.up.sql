CREATE INDEX IF NOT EXISTS titles_title_idx ON titles USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS titles_country_idx ON titles USING GIN (to_tsvector('simple', country));