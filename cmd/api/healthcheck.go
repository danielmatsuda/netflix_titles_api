package main

import (
	"net/http"
)

// healthcheckHandler is implemented as a method on the application struct. This way, needed parameters to the function
// can be included as fields in the application struct.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// create a map that holds health check data
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
