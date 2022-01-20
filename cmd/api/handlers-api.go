package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mklef121/go-card-charge/internal/cards"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Plan          string `json:"plan"`
	LastFour      string `json:"last_four"`
}

type jsonResponse struct {
	OK      bool   "json:\"ok\""
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id,omitempty"`
}

func (app *application) GetPaymentIntent(writer http.ResponseWriter, request *http.Request) {
	var payload stripePayload
	// request.Response.Body
	err := json.NewDecoder(request.Body).Decode(&payload)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Currency: payload.Currency,
		Key:      app.config.stripe.pubKey,
	}

	okay := true

	intent, msg, err := card.Charge(payload.Currency, amount)

	if err != nil {
		okay = false
	}

	if okay {
		out, err := json.MarshalIndent(intent, "", "   ")

		if err != nil {
			app.errorLog.Println(err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")

		writer.Write(out)

	} else {
		response := jsonResponse{
			OK:      false,
			Message: msg,
			Content: "",
		}
		data, err := json.MarshalIndent(response, "", "        ")

		if err != nil {
			app.errorLog.Println(err)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(data)
	}

}

func (app *application) GetWidgetById(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")
	widgetId, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetId)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	out, err := json.MarshalIndent(widget, "", "  ")

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(out)
}

func (app *application) CreateCustomerAndSubscribe(writer http.ResponseWriter, request *http.Request) {
	var payload stripePayload
	// request.Response.Body
	err := json.NewDecoder(request.Body).Decode(&payload)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(payload)

	okay := true
	msg := ""

	respo := jsonResponse{
		OK:      okay,
		Message: msg,
	}

	out, err := json.MarshalIndent(respo, "", "  ")

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(out)

}
