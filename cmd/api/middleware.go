package main

import (
	"expvar"
	"fmt"
	"net/http"
	"strconv"

	"github.com/felixge/httpsnoop"
	"golang.org/x/time/rate"
)

/*
Methods in this file wrap each of my http Handlers. The middleware logic is added to the HandlerFunc's logic,
so it will run for every requested handled by that Handler. The middleware's init logic is run only once per Handler.
*/

// recoverPanic will send a 500 Internal Server Error to the client and close that connection,
// in the event of a panic.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// run the deferred func in the event of a panic
		defer func() {
			if err := recover(); err != nil {
				// trigger Go to close the connection upon responding
				w.Header().Set("Connection", "close")
				// send a 500 Internal Server Error response
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// rateLimit applies the rate limit rules, found in app.config.limiter, to the router.
func (app *application) rateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)
	// return a Go closure. http.HandlerFunc takes in an anonymous function as an argument, which itself checks the rate limit.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// limiter.Allow() checks if the request can be made. If not, return a 429 Too Many Requests response
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) metrics(next http.Handler) http.Handler {
	// publish request-level metrics in expvar
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_microseconds")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	// update the metric totals upon each request
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// before routing the request, count it
		totalRequestsReceived.Add(1)

		// use httpsnoop to capture the metrics of the next handler in the chain
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// on the way back up the middleware's call chain, we know a response has been sent
		totalResponsesSent.Add(1)
		// update the processing time and add the response status code
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
