package main

import "net/http"

func SessionLoad(next http.Handler) http.Handler {
	return sessionManager.LoadAndSave(next)
}

func (app *application) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if !app.session.Exists(r.Context(), "userID") {
			http.Redirect(rw, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(rw, r)
	})
}
