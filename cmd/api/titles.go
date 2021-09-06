package main

import (
	"errors"
	"fmt"
	"net/http"

	"danielmatsuda15.rest/internal/data"
	"danielmatsuda15.rest/internal/validator"
)

// createTitleHandler handles POST requests to the "/v1/titles" endpoint.
// Currently, only allows the client to update one item at a time.
func (app *application) createTitleHandler(w http.ResponseWriter, r *http.Request) {
	// create a struct to hold the POST request body params that we're willing to accept from the client
	var input struct {
		TitleType   string `json:"title_type"`
		Title       string `json:"title"`
		Director    string `json:"director"`
		Country     string `json:"country"`
		ReleaseYear int32  `json:"release_year"`
	}
	// init a new json.Decoder instance, which reads from the request body
	// and uses the Decode() method to dump the relevant key/value pairs into input using pointers
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Copy the values from the input struct to a new Title struct.
	title := &data.Title{
		TitleType:   input.TitleType,
		Title:       input.Title,
		Director:    input.Director,
		Country:     input.Country,
		ReleaseYear: input.ReleaseYear,
	}

	v := validator.New()

	// validate the params and send an error response to the client if necessary
	if data.ValidateTitle(v, title); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert the new title into the db, and write the new item's id to title.ID
	err = app.models.Titles.Insert(title)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// send an HTTP response. Write a Location header with the new item's URL
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/titles/%d", title.ID))
	// write the JSON response
	err = app.writeJSON(w, http.StatusCreated, envelope{"title": title}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showTitleHandler handles GET requests to the "/v1/titles/:id" endpoint.
func (app *application) showTitleHandler(w http.ResponseWriter, r *http.Request) {
	// read the requested id as an int
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// send SQL query for given id. If a data.ErrRecordNotFound error is returned, send 404 response to client
	title, err := app.models.Titles.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// entry found. Write its data as JSON response to client
	err = app.writeJSON(w, http.StatusOK, envelope{"title": title}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateTitleHandler handles PUT requests to the "/v1/titles/:id" endpoint.
func (app *application) updateTitleHandler(w http.ResponseWriter, r *http.Request) {
	// read the requested id as an int
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// call GET on the id to update, to make sure it exists
	title, err := app.models.Titles.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// entry found. Create a struct to hold request data
	var input struct {
		ID          int32  `json:"id"`
		TitleType   string `json:"title_type"`
		Title       string `json:"title"`
		Director    string `json:"director"`
		Country     string `json:"country"`
		ReleaseYear int32  `json:"release_year"`
	}

	// read the JSON response data from the Get() call into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// copy this data into the Title instance and validate the request param values
	title.TitleType = input.TitleType
	title.Title = input.Title
	title.Director = input.Director
	title.Country = input.Country
	title.ReleaseYear = input.ReleaseYear
	title.ID = id

	v := validator.New()
	if data.ValidateTitle(v, title); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// finally, use the request values from title to update the entry
	err = app.models.Titles.Update(title)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated data as a JSON response to client
	err = app.writeJSON(w, http.StatusOK, envelope{"title": title}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteTitleHandler handles DELETE requests on the "/v1/titles/:id" endpoint.
func (app *application) deleteTitleHandler(w http.ResponseWriter, r *http.Request) {
	// read the requested id as an int
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// delete the entry with the given id, or send error to client if entry not found
	err = app.models.Titles.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// DELETE was successful, so write a 200 OK status and message as JSON to client
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "title successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listTitlesHandler handles GET requests to the "/v1/titles" endpoint.
// Sends a JSON response containing all titles from the titles table that match
// the filtering criteria passed in by the client. The filters are passed in by the client as query string parameters.
func (app *application) listTitlesHandler(w http.ResponseWriter, r *http.Request) {
	// get the url.Values map of query string data
	queryString := r.URL.Query()

	// define an input struct to hold possible filter params
	var input struct {
		TitleType string
		Title     string
		Director  string
		Country   string
	}

	// read the parameters into input, possibly using converted param vals, or their defaults if not provided
	input.TitleType = app.readString(queryString, "title_type", "")
	input.Title = app.readString(queryString, "title", "")
	input.Director = app.readString(queryString, "director", "")
	input.Country = app.readString(queryString, "country", "")

	// run a GET request, filtering on these params
	filterParams := []interface{}{input.Title, input.Country, input.TitleType, input.Director}
	titles, err := app.models.Titles.GetAll(filterParams)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// return a JSON response to the client
	err = app.writeJSON(w, http.StatusOK, envelope{"titles": titles}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
