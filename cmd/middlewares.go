package main

import (
	"fmt"
	"net/http"
)

// Go's HTTP server already handle panic in handler. This middeware send InternalErrorResponse
// to client which is better than no response at all.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			pv := recover()
			if pv != nil {
				// setting this header will make Go's HTTP server automatically close
				// the current connection after the response has been sent. We don't set this header on
				// every errorResponse because we want to reuse the connection.
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%v", pv))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
