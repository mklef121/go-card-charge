package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) loadRoutes() http.Handler {

	router := chi.NewRouter()

	router.Get("/virtual-terminal", app.VirtualTerminal)
	router.Post("/paymet-succeeded", app.PaymentSucceeded)

	return router

}
