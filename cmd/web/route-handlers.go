package main

import (
	"net/http"

	"github.com/mklef121/go-card-charge/internal/models"
)

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

	uiData := make(map[string]interface{})
	uiData["cardholder"] = cardHolder
	uiData["email"] = cardHolderEmail
	uiData["paymentIntent"] = paymentIntent
	uiData["paymentMethod"] = paymentMethod
	uiData["paymentCurrency"] = paymentCurrency
	uiData["paymentAmount"] = paymentAmount

	// fmt.Println(uiData)
	_, err = app.renderTemplate(writer, request, "succeeded", &templateData{Data: uiData})
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

//Displays the page to charge one widget
func (app *application) ChargeOnce(writer http.ResponseWriter, request *http.Request) {

	widget := models.Widget{
		ID:             1,
		Name:           "Custom Widget",
		Description:    "A very nice widget",
		InventoryLevel: 10,
		Price:          1000,
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if _, err := app.renderTemplate(writer, request, "buy-once", &templateData{Data: data}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}
