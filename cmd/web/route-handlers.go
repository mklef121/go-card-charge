package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mklef121/go-card-charge/internal/cards"
	"github.com/mklef121/go-card-charge/internal/models"
)

type TransactionData struct {
	FirstName       string
	LastName        string
	Email           string
	PaymentIntentID string
	PaymentMethod   string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

//Gets transaction Data from post info and stripe
func (app *application) GetTransactionData(request *http.Request) (TransactionData, error) {
	var tranxData TransactionData

	// fmt.Println(request.Body)
	err := request.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return tranxData, err
	}

	//read the posted data
	cardHolderEmail := request.Form.Get("cardholder_email")
	firstName := request.Form.Get("first_name")
	lastName := request.Form.Get("last_name")
	paymentIntent := request.Form.Get("payment_intent")
	paymentMethod := request.Form.Get("payment_method")
	paymentAmount := request.Form.Get("payment_amount")
	paymentCurrency := request.Form.Get("payment_currency")
	amount, _ := strconv.Atoi(paymentAmount)

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.pubKey,
	}

	pi, err := card.RetrieveExistingPaymentIntent(paymentIntent)

	if err != nil {
		app.errorLog.Println(err)
		return tranxData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)

	if err != nil {
		app.errorLog.Println(err)
		return tranxData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	tranxData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           cardHolderEmail,
		PaymentIntentID: paymentMethod,
		PaymentMethod:   paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.Charges.Data[0].ID,
	}

	return tranxData, nil

}
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

func (app *application) VirtualTerminalPaymentSucceeded(writer http.ResponseWriter, request *http.Request) {

	reqData, err := app.GetTransactionData(request)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	transaction := models.Transaction{
		Amount:              reqData.PaymentAmount,
		Currency:            reqData.PaymentCurrency,
		LastFour:            reqData.LastFour,
		ExpiryMonth:         reqData.ExpiryMonth,
		ExpiryYear:          reqData.ExpiryYear,
		BankReturnCode:      reqData.BankReturnCode,
		TransactionStatusID: 2,
		PaymentIntent:       reqData.PaymentIntentID,
		PaymentMethod:       reqData.PaymentMethod,
	}

	_, err = app.SaveTransaction(transaction)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.session.Put(request.Context(), "virtual-terminal-receipt", reqData)

	http.Redirect(writer, request, "/virtual-terminal-receipt", http.StatusSeeOther)

}

func (app *application) VirtualTerminalPaymentReciept(writer http.ResponseWriter, request *http.Request) {
	transXn := app.session.Get(request.Context(), "virtual-terminal-receipt").(TransactionData)
	templData := make(map[string]interface{})
	templData["txn"] = transXn
	// app.session.Remove(request.Context(), "receipt")
	_, err := app.renderTemplate(writer, request, "virtual-terminal-receipt", &templateData{Data: templData})
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

func (app *application) PaymentSucceeded(writer http.ResponseWriter, request *http.Request) {

	reqData, err := app.GetTransactionData(request)

	if err != nil {
		app.errorLog.Println(err)
		return
	}
	productID, _ := strconv.Atoi(request.Form.Get("product_id"))

	//Create a new customer

	customerID, err := app.SaveCustomer(reqData.FirstName, reqData.LastName, reqData.Email)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create a new transaction

	transaction := models.Transaction{
		Amount:              reqData.PaymentAmount,
		Currency:            reqData.PaymentCurrency,
		LastFour:            reqData.LastFour,
		ExpiryMonth:         reqData.ExpiryMonth,
		ExpiryYear:          reqData.ExpiryYear,
		BankReturnCode:      reqData.BankReturnCode,
		TransactionStatusID: 2,
		PaymentIntent:       reqData.PaymentIntentID,
		PaymentMethod:       reqData.PaymentMethod,
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
		Amount:        reqData.PaymentAmount,
	}

	_, err = app.SaveOrder(order)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.session.Put(request.Context(), "receipt", reqData)

	http.Redirect(writer, request, "/receipt", http.StatusSeeOther)
}

func (app *application) Receipt(writer http.ResponseWriter, request *http.Request) {
	transXn := app.session.Get(request.Context(), "receipt").(TransactionData)
	templData := make(map[string]interface{})
	templData["txn"] = transXn
	// app.session.Remove(request.Context(), "receipt")
	_, err := app.renderTemplate(writer, request, "receipt", &templateData{Data: templData})
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

func (app *application) GoldPlanReceipt(writer http.ResponseWriter, request *http.Request) {

	_, err := app.renderTemplate(writer, request, "receipt-plan", nil)
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

func (app *application) GoldPlan(writer http.ResponseWriter, request *http.Request) {

	widget, err := app.DB.GetWidget(2)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	dataMap := make(map[string]interface{})

	dataMap["widget"] = widget

	_, err = app.renderTemplate(writer, request, "gold-plan", &templateData{Data: dataMap})
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

func (app *application) LoginPage(writer http.ResponseWriter, request *http.Request) {

	_, err := app.renderTemplate(writer, request, "login", nil)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}
