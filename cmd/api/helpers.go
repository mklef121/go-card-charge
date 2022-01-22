package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/mklef121/go-card-charge/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type ApiMessage struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Token   *models.Token `json:"token,omitempty"`
}

func (app *application) readJson(writer http.ResponseWriter, request *http.Request, data interface{}) error {
	var maxBytes int64 = 1048576

	request.Body = http.MaxBytesReader(writer, request.Body, maxBytes)

	dec := json.NewDecoder(request.Body)
	err := dec.Decode(data)

	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("Body must only have a single JSON value")
	}

	return nil
}

func (app *application) badRequest(writer http.ResponseWriter, err error) error {
	var payload ApiMessage

	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")

	if err != nil {
		return err
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write(out)

	return nil
}

//Writes data out as Json
func (app *application) writeJson(writer http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {

	out, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	// if len(headers) > 0 {

	// for k,value := range headers {
	// 	writer.Header()

	// }
	// }

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	writer.Write(out)

	return nil
}

func (app *application) invalidCredentials(writer http.ResponseWriter) error {

	payload := ApiMessage{
		Error:   true,
		Message: "invalid Authentication credentials",
	}

	err := app.writeJson(writer, http.StatusUnauthorized, payload)

	if err != nil {
		return err
	}
	return nil
}

func (app *application) passwordMatches(password, hash string) (bool, error) {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
