package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) loadRoutes() http.Handler {

	router := chi.NewRouter()

	router.Get("/virtual-terminal", app.VirtualTerminal)
	router.Post("/paymet-succeeded", app.PaymentSucceeded)
	router.Get("/charge-once", app.ChargeOnce)

	fileServer := http.FileServer(http.Dir("./static"))

	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return router

}
