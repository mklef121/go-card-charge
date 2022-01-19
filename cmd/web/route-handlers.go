package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mklef121/go-card-charge/internal/cards"
)

func (app *application) HomePage(writer http.ResponseWriter, r *http.Request) {

	if _, err := app.renderTemplate(writer, r, "home", &templateData{}); err != nil {

		app.errorLog.Println(err)
	}

}

func (app *application) VirtualTerminal(writer http.ResponseWriter, r *http.Request) {

	if _, err := app.renderTemplate(writer, r, "terminal", &templateData{}, "stripe-js"); err != nil {

		app.errorLog.Println(err)
	}

}

func (app *application) PaymentSucceeded(writer http.ResponseWriter, request *http.Request) {

	// fmt.Println(request.Body)
	err := request.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//read the posted data
	cardHolder := request.Form.Get("cardholder_name")
	cardHolderEmail := request.Form.Get("cardholder_email")
	paymentIntent := request.Form.Get("payment_intent")
	paymentMethod := request.Form.Get("payment_method")
	paymentAmount := request.Form.Get("payment_amount")
	paymentCurrency := request.Form.Get("payment_currency")

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.pubKey,
	}

	pi, err := card.RetrieveExistingPaymentIntent(paymentIntent)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	pm, err := card.GetPaymentMethod(paymentMethod)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	//Create a new customer
	// create a new transaction
	// create a new order
	uiData := make(map[string]interface{})
	uiData["cardholder"] = cardHolder
	uiData["email"] = cardHolderEmail
	uiData["paymentIntent"] = paymentIntent
	uiData["paymentMethod"] = paymentMethod
	uiData["paymentCurrency"] = paymentCurrency
	uiData["paymentAmount"] = paymentAmount
	uiData["lastFour"] = lastFour
	uiData["expiryMonth"] = expiryMonth
	uiData["expiryYear"] = expiryYear
	uiData["bankReturnCode"] = pi.Charges.Data[0].ID

	// fmt.Println(uiData)
	_, err = app.renderTemplate(writer, request, "succeeded", &templateData{Data: uiData})
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

//Displays the page to charge one widget
func (app *application) ChargeOnce(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")
	widgetId, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetId)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if _, err := app.renderTemplate(writer, request, "buy-once", &templateData{Data: data}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}
