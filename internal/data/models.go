package data

import (
	"database/sql"
	"errors"
)

// make a custom error, returned from Get() when looking up an item that doesn't exist
var (
	ErrRecordNotFound = errors.New("record not found")
)

// Models wraps all database models, so they can be found in one place
type Models struct {
	Titles TitleModel
}

// NewModels constructs a new Model
func NewModels(db *sql.DB) Models {
	return Models{
		Titles: TitleModel{DB: db},
	}
}
