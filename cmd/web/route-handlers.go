package main

import "net/http"

func (app *application) VirtualTerminal(writer http.ResponseWriter, r *http.Request) {

	if _, err := app.renderTemplate(writer, r, "terminal", nil); err != nil {

		app.errorLog.Println(err)

	}

}
