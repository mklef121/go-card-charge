package cards

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	Amount              int
	TransactionStatusId int
	Currency            string
	LastFourDigit       string
	BankReturnCode      string
}

func (card *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return card.CreatePaymentIntent(currency, amount)
}

func (card *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = card.Secret
	// fmt.Println(card)

	//create payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	// params.AddMetadata("key", "value")
	pi, err := paymentintent.New(params)

	if err != nil {
		msg := ""
		if stripeError, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeError.Code)
		}

		return nil, msg, err
	}

	return pi, "", nil
}

func cardErrorMessage(code stripe.ErrorCode) string {
	var msg = ""
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined"
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card is expired"
	case stripe.ErrorCodeIncorrectCVC:
		msg = "Incorrect CVC code"
	case stripe.ErrorCodeIncorrectZip:
		msg = "Incorrect zip/postal code"
	case stripe.ErrorCodeAmountTooLarge:
		msg = "The amount is too large to charge to your card"
	case stripe.ErrorCodeAmountTooSmall:
		msg = "The amount is too small to charge to your card"
	case stripe.ErrorCodeBalanceInsufficient:
		msg = "Insufficient balance"
	case stripe.ErrorCodePostalCodeInvalid:
		msg = "Your postal code is invalid"
	default:
		msg = "Your card was declined"
	}
	return msg
}

//Gets the payment method by payment intent id
func (card *Card) GetPaymentMethod(id string) (*stripe.PaymentMethod, error) {
	stripe.Key = card.Secret

	pm, err := paymentmethod.Get(id, nil)

	if err != nil {
		return nil, err
	}

	return pm, nil
}

//Gets an existing payment intent by ID
func (card *Card) RetrieveExistingPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	stripe.Key = card.Secret

	pi, err := paymentintent.Get(id, nil)

	if err != nil {
		return nil, err
	}

	return pi, nil
}
