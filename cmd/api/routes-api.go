package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) loadApiRoutes() http.Handler {

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Route("/api", func(chiRouter chi.Router) {
		chiRouter.Post("/payment-intent", app.GetPaymentIntent)
		chiRouter.Get("/widgets/{id}", app.GetWidgetById)
		chiRouter.Post("/create-customer-and-subscribe", app.CreateCustomerAndSubscribe)
		chiRouter.Post("/authenticate", app.AuthenticateUser)
		chiRouter.Post("/is-authenticated", app.CheckAuthentication)

		chiRouter.Route("/admin", func(r chi.Router) {
			r.Use(app.Auth)

			r.Get("/test", func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("Just got in"))
			})

			r.Post("/virtual-terminal-succeeded", app.VirtualTerminalPaymentSuccess)
		})

	})

	router.Get("/me-u", app.GetPaymentIntent)

	return router

}
