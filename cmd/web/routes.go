package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) loadRoutes() http.Handler {

	router := chi.NewRouter()

	router.Use(SessionLoad)

	router.Route("/admin", func(childRouter chi.Router) {
		childRouter.Use(app.Auth)

		childRouter.Get("/virtual-terminal", app.VirtualTerminal)

	})
	router.Get("/", app.HomePage)

	// router.Get("/virtual-terminal-receipt", app.VirtualTerminalPaymentReciept)

	// router.Post("/virtual-terminal-payment-succeeded", app.VirtualTerminalPaymentSucceeded)

	router.Post("/payment-succeeded", app.PaymentSucceeded)
	router.Get("/widgets/{id}/charge-once", app.ChargeOnce)
	router.Get("/receipt", app.Receipt)

	router.Get("/plans/gold", app.GoldPlan)
	router.Get("/receipt/gold-plan", app.GoldPlanReceipt)

	//Auth pages
	router.Get("/login", app.LoginPage)
	router.Post("/login", app.PostLoginPage)
	router.Get("/logout", app.Logout)
	router.Get("/forgot-password", app.ForgotPassword)

	fileServer := http.FileServer(http.Dir("./static"))

	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return router

}
