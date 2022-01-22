package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mklef121/go-card-charge/internal/cards"
	"github.com/mklef121/go-card-charge/internal/models"
	"github.com/stripe/stripe-go/v72"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Plan          string `json:"plan"`
	LastFour      string `json:"last_four"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int    `json:"expiry_month"`
	ExpiryYear    int    `json:"expiry_year"`
	ProductID     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
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

	respo := jsonResponse{
		OK:      false,
		Message: "",
	}
	// request.Response.Body
	err := json.NewDecoder(request.Body).Decode(&payload)

	if err != nil {
		app.errorLog.Println(err)
		app.writeJsonResponse(respo, writer, err)
		return
	}

	app.infoLog.Println(payload)

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Currency: payload.Currency,
		Key:      app.config.stripe.pubKey,
	}

	var subscription *stripe.Subscription

	stripeCustomer, msg, err := card.CreateCustomer(payload.PaymentMethod, payload.Email)

	if err != nil {
		app.errorLog.Println(err)

		if msg == "" {
			msg = err.Error()
		}

		respo.Message = msg
		app.writeJsonResponse(respo, writer, nil)
		return
	}

	subscription, err = card.SubscribeToPlan(stripeCustomer, payload.Plan, payload.Email, payload.LastFour, "")

	app.infoLog.Println("The subscription ID is", subscription.ID)

	if err != nil {
		app.errorLog.Println(err)
		app.writeJsonResponse(respo, writer, err)
		return
	}

	productId, _ := strconv.Atoi(payload.ProductID)
	customerID, err := app.SaveCustomer(payload.FirstName, payload.LastName, payload.Email)

	if err != nil {
		app.errorLog.Println(err)
		app.writeJsonResponse(respo, writer, err)
		return
	}

	//Save a transaction
	amount, _ := strconv.Atoi(payload.Amount)

	transaction := models.Transaction{
		Amount:      amount,
		Currency:    "cad",
		LastFour:    payload.LastFour,
		ExpiryMonth: payload.ExpiryMonth,
		ExpiryYear:  payload.ExpiryYear,
		// BankReturnCode:      payload.r,
		TransactionStatusID: 2,
		// PaymentIntent:       payload.int,
		PaymentMethod: payload.PaymentMethod,
	}

	trxnId, err := app.SaveTransaction(transaction)

	if err != nil {
		app.errorLog.Println(err)
		app.writeJsonResponse(respo, writer, err)
		return
	}

	order := models.Order{
		WidgetID:      productId,
		TransactionID: trxnId,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        amount,
	}

	_, err = app.SaveOrder(order)

	if err != nil {
		app.errorLog.Println(err)
		app.writeJsonResponse(respo, writer, err)
		return
	}

	respo.OK = true
	respo.Message = "Successfully subscribed to plan"
	app.writeJsonResponse(respo, writer, nil)

}

func (app *application) writeJsonResponse(resp jsonResponse, writer http.ResponseWriter, err error) {
	if err != nil {
		resp.Message = err.Error()
	}
	out, err := json.MarshalIndent(resp, "", "  ")

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(out)
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

func (app *application) SaveOrder(order models.Order) (int, error) {

	id, err := app.DB.InsertOrder(order)

	if err != nil {
		return 0, err
	}

	return id, nil
}
