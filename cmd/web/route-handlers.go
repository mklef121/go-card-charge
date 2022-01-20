package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mklef121/go-card-charge/internal/cards"
	"github.com/mklef121/go-card-charge/internal/models"
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
	cardHolderEmail := request.Form.Get("cardholder_email")
	firstName := request.Form.Get("first_name")
	lastName := request.Form.Get("last_name")
	paymentIntent := request.Form.Get("payment_intent")
	paymentMethod := request.Form.Get("payment_method")
	paymentAmount := request.Form.Get("payment_amount")
	paymentCurrency := request.Form.Get("payment_currency")
	productID, _ := strconv.Atoi(request.Form.Get("product_id"))

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

	customerID, err := app.SaveCustomer(firstName, lastName, cardHolderEmail)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create a new transaction

	amount, _ := strconv.Atoi(paymentAmount)
	transaction := models.Transaction{
		Amount:              amount,
		Currency:            paymentCurrency,
		LastFour:            lastFour,
		ExpiryMonth:         int(expiryMonth),
		ExpiryYear:          int(expiryYear),
		BankReturnCode:      pi.Charges.Data[0].ID,
		TransactionStatusID: 2,
		PaymentIntent:       paymentIntent,
		PaymentMethod:       paymentMethod,
	}

	trxID, err := app.SaveTransaction(transaction)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create a new order
	order := models.Order{
		WidgetID:      productID,
		CustomerID:    customerID,
		TransactionID: trxID,
		StatusID:      1,
		Quantity:      1,
		Amount:        amount,
	}

	_, err = app.SaveOrder(order)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	uiData := make(map[string]interface{})
	uiData["firstName"] = firstName
	uiData["lastName"] = lastName
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

//saves a customer and returns ID
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := app.DB.InsertCustomer(customer)

	if err != nil {
		return 0, err
	}

	return id, nil
}

//saves a a transaction to db and returns ID
func (app *application) SaveTransaction(trnx models.Transaction) (int, error) {

	id, err := app.DB.InsertTransaction(trnx)

	if err != nil {
		return 0, err
	}

	return id, nil
}

//saves a a transaction to db and returns ID
func (app *application) SaveOrder(order models.Order) (int, error) {

	id, err := app.DB.InsertOrder(order)

	if err != nil {
		return 0, err
	}

	return id, nil
}
