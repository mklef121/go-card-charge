package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"
const cssVersion = "1"

type Config struct {
	port int
	env  string
	api  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		pubKey string
	}
}

type application struct {
	config        Config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	version       string
}

func (app *application) serve() error {
	server := http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting HTTP server in %s mode on %d port \n", app.config.env, app.config.port)

	return server.ListenAndServe()
}

func main() {
	var appConfig Config
	flag.IntVar(&appConfig.port, "port", 4000, "Server port to listen")
	flag.StringVar(&appConfig.env, "env", "development", "application development environment {development | production}")
	flag.StringVar(&appConfig.api, "api", "http://localhost:4001", "URL to api")

	flag.Parse()

	appConfig.stripe.pubKey = os.Getenv("STRIPE_KEY")
	appConfig.stripe.secret = os.Getenv("STRIPE_SECRET_KEY")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	tc := make(map[string]*template.Template)

	app := application{
		config:        appConfig,
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: tc,
		version:       version,
	}

	err := app.serve()

	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
