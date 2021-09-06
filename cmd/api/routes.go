package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes returns an http.Handler that routes API requests. Consists of an httprouter wrapped in middleware.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// customize some router error responses
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// add routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/titles", app.listTitlesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/titles", app.createTitleHandler)

	router.HandlerFunc(http.MethodGet, "/v1/titles/:id", app.showTitleHandler)
	router.HandlerFunc(http.MethodPut, "/v1/titles/:id", app.updateTitleHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/titles/:id", app.deleteTitleHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// apply middleware logic before any actual routing occurs
	return app.metrics(app.recoverPanic(app.rateLimit(router)))
}
