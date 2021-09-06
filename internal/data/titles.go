package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"danielmatsuda15.rest/internal/validator"
)

// Title holds values parsed from the client's POST request body.
type Title struct {
	ID          int64  `json:"id"`
	TitleType   string `json:"title_type"`
	Title       string `json:"title"`
	Director    string `json:"director"`
	Country     string `json:"country"`
	ReleaseYear int32  `json:"release_year"`
}

// ValidateTitle validates the client's request params according to my API's business logic/rules. Takes in an
// empty Validator instance, and a Title struct containing values from the client.
func ValidateTitle(v *validator.Validator, title *Title) {
	// Use the Check() method to execute validation checks
	v.Check(title.TitleType != "", "title_type", "must be provided")
	v.Check(title.Title != "", "title", "must be provided")
	v.Check(title.Director != "", "director", "must be provided")
	v.Check(title.Country != "", "country", "must be provided")

	v.Check(title.ReleaseYear != 0, "release_year", "must be provided")
	v.Check(title.ReleaseYear >= 1888, "release_year", "must be greater than 1888")
	v.Check(title.ReleaseYear <= int32(time.Now().Year()), "release_year", "must not be in the future")
}

type TitleModel struct {
	DB *sql.DB
}

// Insert inserts a new row into the titles table.
// It takes a pointer to a Title struct. That Title contains the data to populate the new record.
func (t TitleModel) Insert(title *Title) error {
	// create new entry and return some data for the API's response
	query := `
	INSERT INTO titles (title_type, title, director, country, release_year)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`

	// args to pass into SQL placeholders. If necessary, convert types here using pq
	args := []interface{}{title.TitleType, title.Title, title.Director, title.Country, title.ReleaseYear}

	// create an empty context.Context instance, with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// release context's resources before Get() returns. Otherwise, those resources will be held
	// until timeout or the parent context is canceled. Prevents memory leaks!
	defer cancel()

	// QueryRow() executes query on the t.DB connection pool, with args as a variadic param.
	// Scan writes the query's returned values into fields of the title struct (here, just title.ID)
	return t.DB.QueryRowContext(ctx, query, args...).Scan(&title.ID)
}

// Get uses the id parameter given to return a single row from the db in a Title struct instance.
// May return an ErrRecordNotFound error if the query is invalid.
func (t TitleModel) Get(id int64) (*Title, error) {
	// additional naive check for valid id
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
	SELECT *
	FROM titles
	WHERE id = $1`

	// hold the returned data in a new Title struct
	var title Title

	// create an empty context.Context instance, with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// release context's resources before Get() returns. Otherwise, those resources will be held
	// until timeout or the parent context is canceled. Prevents memory leaks!
	defer cancel()

	// Execute query and scan response into Title struct.
	// Scan writes the query's returned values into specified fields of the title struct
	err := t.DB.QueryRowContext(ctx, query, id).Scan(
		&title.ID,
		&title.TitleType,
		&title.Title,
		&title.Director,
		&title.Country,
		&title.ReleaseYear,
	)
	// If no matching entry was found, a sql.ErrNoRows error will be returned.
	// In this case, return the custom ErrRecordNotFound error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// if operation is successful, return a pointer to the Title struct
	return &title, nil
}

// GetAll returns all rows and all columns from the titles table in a slice -- if the filterParams are all
// empty strings. Otherwise, only returns the slice of all rows and columns that meet filter criteria.
func (t TitleModel) GetAll(filterParams []interface{}) ([]*Title, error) {
	// each filter param's WHERE clause is skipped if the value passed in is an empty string
	// PostgreSQL allows for full text searches:
	// https://www.postgresql.org/docs/13/textsearch-intro.html#TEXTSEARCH-MATCHING
	query := `
	SELECT *
	FROM titles
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple', country) @@ plainto_tsquery('simple', $2) OR $2 = '')
	AND (LOWER(title_type) = LOWER($3) OR $3 = '')
	AND (LOWER(director) = LOWER($4) OR $4 = '')
	ORDER BY id`

	// create an empty context.Context instance, with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// release context's resources before Get() returns. Otherwise, those resources will be held
	// until timeout or the parent context is canceled. Prevents memory leaks!
	defer cancel()

	// execute the query. Returns a sql.Results object into the rows var
	rows, err := t.DB.QueryContext(ctx, query, filterParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// copy each result row into its own Title struct. Hold all Titles in a slice.
	titles := []*Title{}
	for rows.Next() {
		var title Title
		err := rows.Scan(
			&title.ID,
			&title.TitleType,
			&title.Title,
			&title.Director,
			&title.Country,
			&title.ReleaseYear,
		)
		if err != nil {
			return nil, err
		}
		titles = append(titles, &title)
	}

	// confirm there were no errors during the calls to row.Next()
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// finally, return the slice of titles if no errors were found
	return titles, nil
}

// Update runs a SQL UPDATE command using data params from title. If successful,
// it returns the entry's updated data in the title struct.
func (t TitleModel) Update(title *Title) error {
	// to update an entry, you must provide ALL values, including values that haven't changed
	query := `
UPDATE titles
SET title_type = $1, title = $2, director = $3, country = $4, release_year = $5
WHERE id = $6
RETURNING id, title_type, title, director, country, release_year`

	// params from title to pass into query
	args := []interface{}{
		title.TitleType,
		title.Title,
		title.Director,
		title.Country,
		title.ReleaseYear,
		title.ID,
	}

	// create an empty context.Context instance, with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// release context's resources before Get() returns. Otherwise, those resources will be held
	// until timeout or the parent context is canceled. Prevents memory leaks!
	defer cancel()

	// query and read the result into title. Then return title
	return t.DB.QueryRowContext(ctx, query, args...).Scan(
		&title.ID,
		&title.TitleType,
		&title.Title,
		&title.Director,
		&title.Country,
		&title.ReleaseYear)
}

// Delete deletes the entry with the given id, and returns nil if successful. If the entry with that id
// doesn't exist in the database, returns an error.
func (t TitleModel) Delete(id int64) error {
	// additional naive check for valid id
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM titles
	WHERE id = $1`

	// create an empty context.Context instance, with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// release context's resources before Get() returns. Otherwise, those resources will be held
	// until timeout or the parent context is canceled. Prevents memory leaks!
	defer cancel()

	// execute the query for the given id. Returns a sql.Result object
	result, err := t.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// find the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// if no rows were affected, no DELETE operation took place (the id didn't exist), so
	// return a not found error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
