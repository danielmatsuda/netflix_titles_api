package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// readIDParam extracts the string value of the id parameter, then tries to convert it to int.
// If this fails or the id is less than 1, the ID is invalid, so returns 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// grab the slice of parameter names/values from URL params
	params := httprouter.ParamsFromContext(r.Context())

	id := params.ByName("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil || idInt < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return idInt, nil
}

// used to envelop the JSON response before it is sent
type envelope map[string]interface{}

// writeJSON writes data as JSON to stdout. Takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// MarshalIndent() prefixes the data with "", and indents with "\t"; returns the data as a JSONified byte slice.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	js = append(js, '\n')

	// write each header from the headers map to the ResponseWriter's headers
	for key, value := range headers {
		w.Header()[key] = value
	}

	// set the Content-Type header to indicate JSON is being sent, and write the response code status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// write response body
	w.Write(js)

	return nil
}

// readJSON reads JSON from a POST request into the dst interface. The method provides error handling
// for invalid requests, limits the max request body size, disallows unknown fields, and
// allows only one JSON object in the request body.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use http.MaxBytesReader() to limit the req body size to maxBytes only
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// disallow unknown fields when decoding request JSON
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body to the destination
	err := dec.Decode(dst)
	if err != nil {
		// handle client errors
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format "json: unknown
		// field "<name>"".
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			// in this case, extract the field name from the error,
			// and interpolate it into our custom error message.
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// If the request body exceeds 1MB in size, the decode will fail
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	// to ensure the request contains only one JSON object, decode again and cause an io.EOF error.
	// otherwise, return a custom error message
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// readString returns a string value from the query string, or the defaultValue if no matching key could be found.
func (app *application) readString(queryString url.Values, key string, defaultValue string) string {
	// extract the value from the query string for the given key. Returns "" if not found
	stringVal := queryString.Get(key)
	if stringVal == "" {
		return defaultValue
	}
	return stringVal
}
