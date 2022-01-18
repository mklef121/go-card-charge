package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
)

type templateData struct {
	StringMap            map[string]string
	IntMap               map[string]int
	FloatMap             map[string]float32
	Data                 map[string]interface{}
	CSRFToken            string
	Flash                string
	Warning              string
	Error                string
	IsAuthenticated      string
	API                  string
	CSSVersion           string
	StripeSecretKey      string
	StripePublishableKey string
}

var templateFunctions = template.FuncMap{
	"formatCurrency": formatCurrency,
}

func formatCurrency(n int) string {
	f := float32(n / 100)
	return fmt.Sprintf("$%.2f", f)
}

//go:embed templates
var templatesFs embed.FS

//Adds all the default data neded for initial template rendering
func (app *application) addDefaultTemplateData(td *templateData, request *http.Request) *templateData {
	td.API = app.config.api
	td.StripePublishableKey = app.config.stripe.pubKey
	td.StripeSecretKey = app.config.stripe.secret
	return td
}

func (app *application) renderTemplate(writer http.ResponseWriter, request *http.Request, page string, td *templateData, partials ...string) (*template.Template, error) {
	var templ *template.Template
	var err error
	var templateToRender = fmt.Sprintf("templates/%s.page.html", page)

	cacheTempl, templExists := app.templateCache[templateToRender]

	if app.config.env == "production" && templExists {
		templ = cacheTempl
	} else {
		templ, err = app.parseTemplateAndPartials(partials, page, templateToRender)

		if err != nil {
			app.errorLog.Println(err)
			return nil, err
		}
	}

	if td == nil {
		td = &templateData{}
	}

	app.addDefaultTemplateData(td, request)

	err = templ.Execute(writer, td)

	if err != nil {
		app.errorLog.Println(err)

		return nil, err
	}
	return templ, nil
}

func (app *application) parseTemplateAndPartials(partials []string, page string, templPath string) (*template.Template, error) {
	var templ *template.Template
	var err error

	if len(partials) > 0 {
		for index, data := range partials {
			partials[index] = fmt.Sprintf("templates/%s.partial.html", data)
		}
	}

	var parseTmpls = []string{"templates/base.layout.html", templPath}
	if len(partials) > 0 {
		parseTmpls = append(parseTmpls, partials...)
		templ, err = template.New(fmt.Sprintf("%s.page.html", page)).Funcs(templateFunctions).ParseFS(templatesFs, parseTmpls...)
	} else {
		templ, err = template.New(fmt.Sprintf("%s.page.html", page)).Funcs(templateFunctions).ParseFS(templatesFs, parseTmpls...)
	}

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	app.templateCache[templPath] = templ
	return templ, nil
}
