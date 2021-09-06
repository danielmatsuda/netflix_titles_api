CREATE TABLE IF NOT EXISTS titles (
id bigserial PRIMARY KEY,
title_type text NOT NULL,
title text NOT NULL,
director text NOT NULL DEFAULT 'Unknown',
country text NOT NULL DEFAULT 'Unknown',
release_year integer NOT NULL
);