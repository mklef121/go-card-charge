package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

func (app *application) AuthenticateUser(writer http.ResponseWriter, request *http.Request) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJson(writer, request, &userInput)

	if err != nil {
		app.badRequest(writer, err)
		return
	}

	//Get user from db, send error if email is invalid
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		app.invalidCredentials(writer)
		return
	}

	//validate password

	validPass, err := app.passwordMatches(userInput.Password, user.Password)

	if err != nil || !validPass {
		app.invalidCredentials(writer)
		return
	}

	//generate token

	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)

	if err != nil {
		app.badRequest(writer, err)
		return
	}

	err = app.DB.InsertToken(token, user)

	if err != nil {
		app.badRequest(writer, err)
		return
	}
	//send response

	var payload ApiMessage

	payload.Error = false
	payload.Message = fmt.Sprintf("token for %s created", user.Email)
	payload.Token = token

	fmt.Println("The token", token, string(token.PlainText))

	app.writeJson(writer, http.StatusOK, payload)
}

func (app *application) authenticateToken(request *http.Request) (*models.User, error) {

	authHeader := request.Header.Get("Authorization")

	fmt.Println(authHeader, "The authheader")

	if authHeader == "" {
		return nil, errors.New("No authorization header received")
	}

	headParts := strings.Split(authHeader, " ")

	if len(headParts) != 2 || headParts[0] != "Bearer" {
		return nil, errors.New("No authorization header received.")
	}

	token := headParts[1]

	if len(token) != 26 {
		return nil, errors.New("Wrong size of authentication token.")
	}

	//Get the user from the tokens table
	dbUser, err := app.DB.GetUserWithToken(token)

	if err != nil {
		return nil, errors.New("No matching user found.")
	}
	return dbUser, nil
}

func (app *application) CheckAuthentication(writer http.ResponseWriter, request *http.Request) {
	user, err := app.authenticateToken(request)
	if err != nil {
		app.invalidCredentials(writer)
		return
	}

	var res ApiMessage

	res.Error = false
	res.Message = fmt.Sprintf("Authenticated user %s", user.Email)

	app.writeJson(writer, http.StatusOK, res)

}

func (app *application) VirtualTerminalPaymentSuccess(writer http.ResponseWriter, request *http.Request) {
	var txnData struct {
		PaymentAmount   int    `json:"amount"`
		PaymentCurrency string `json:"currency"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		Email           string `json:"email"`
		PaymentIntent   string `json:"payment_intent"`
		PaymentMethod   string `json:"payment_method"`
		BankReturnCode  string `json:"bank_return_code"`
		ExpiryMonth     int    `json:"expiry_month"`
		ExpiryYear      int    `json:"expiry_year"`
		LastFour        string `json:"last_four"`
	}

	err := app.readJson(writer, request, &txnData)

	if err != nil {
		app.badRequest(writer, err)
		return
	}

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.pubKey,
	}

	pi, err := card.RetrieveExistingPaymentIntent(txnData.PaymentIntent)

	if err != nil {
		app.badRequest(writer, err)
		return
	}

	pm, err := card.GetPaymentMethod(txnData.PaymentMethod)

	if err != nil {
		app.badRequest(writer, err)
		return
	}

	txnData.LastFour = pm.Card.Last4
	txnData.ExpiryMonth = int(pm.Card.ExpMonth)
	txnData.ExpiryYear = int(pm.Card.ExpYear)

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		PaymentIntent:       txnData.PaymentIntent,
		BankReturnCode:      pi.Charges.Data[0].ID,
		TransactionStatusID: 2,
		PaymentMethod:       txnData.PaymentMethod,
	}

	_, err = app.SaveTransaction(txn)

	if err != nil {
		app.badRequest(writer, err)
		return
	}

	app.writeJson(writer, http.StatusOK, txn)
}

func (app *application) SendPasswordResetEmail(writer http.ResponseWriter, request *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	err := app.readJson(writer, request, &payload)

	if err != nil {
		app.badRequest(writer, err)
		return
	}

	var data struct {
		Link string
	}

	data.Link = "http://example.com"

	//send email
	err = app.SendMail("info@widget.com", "info@widget.com", "Password Reset", "password-reset", data)

	if err != nil {

		app.infoLog.Println(err, "Got feedback")
		app.badRequest(writer, err)
		return
	}

	var res ApiMessage

	res.Error = false
	res.Message = "Password reset email sent successfully "

	app.writeJson(writer, http.StatusCreated, res)

}
