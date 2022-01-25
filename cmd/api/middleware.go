package main

import "net/http"

func (app *application) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := app.authenticateToken(r)

		if err != nil {
			app.invalidCredentials(rw)
			return
		}

		next.ServeHTTP(rw, r)
	})
}
