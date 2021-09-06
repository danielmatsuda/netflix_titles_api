ALTER TABLE titles ADD CONSTRAINT titles_year_check CHECK (release_year BETWEEN 1888 AND date_part('year', now()));
