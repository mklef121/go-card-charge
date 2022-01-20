package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) loadRoutes() http.Handler {

	router := chi.NewRouter()

	router.Use(SessionLoad)

	router.Get("/", app.HomePage)
	router.Get("/virtual-terminal", app.VirtualTerminal)
	router.Get("/virtual-terminal-receipt", app.VirtualTerminalPaymentReciept)

	router.Post("/virtual-terminal-payment-succeeded", app.VirtualTerminalPaymentSucceeded)

	router.Post("/payment-succeeded", app.PaymentSucceeded)
	router.Get("/widgets/{id}/charge-once", app.ChargeOnce)
	router.Get("/receipt", app.Receipt)

	fileServer := http.FileServer(http.Dir("./static"))

	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return router

}
