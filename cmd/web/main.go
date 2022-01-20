package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/mklef121/go-card-charge/driver"
	"github.com/mklef121/go-card-charge/internal/models"
)

const version = "1.0.0"
const cssVersion = "1"

var sessionManager *scs.SessionManager

type Config struct {
	// app port
	port int
	//wether it is production or development
	env string
	api string
	// Database connection details
	db struct {
		dsn string
	}

	// Stripe account details
	stripe struct {
		secret string
		pubKey string
	}
}

// This holds the entire application data, methods that powers rendering and app functionality
type application struct {
	//Handles access to application config
	config Config
	// Information logging
	infoLog *log.Logger
	// error logging
	errorLog *log.Logger

	//Used to store templates that have been built
	templateCache map[string]*template.Template
	version       string
	DB            models.DBModel
	session       *scs.SessionManager
}

func (app *application) serve() error {
	// normall to server app, we would do
	/*
		http.HandleFunc("/", nil)
		http.ListenAndServe(":8060", nil)
	*/

	// but then http package provides more customizable option

	serverConfig := http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		Handler:           app.loadRoutes(),
	}

	app.infoLog.Printf("Starting HTTP server in %s mode on %d port \n", app.config.env, app.config.port)

	return serverConfig.ListenAndServe()
}

func main() {
	gob.Register(TransactionData{})
	var appConfig Config
	defaultPort := 4000
	defaultApiPort := 4001
	flag.IntVar(&appConfig.port, "port", defaultPort, "Server port to listen")
	flag.StringVar(&appConfig.env, "env", "development", "application development environment {development | production}")
	flag.StringVar(&appConfig.api, "api", "http://localhost:"+strconv.Itoa(defaultApiPort)+"/api", "URL to api")

	flag.Parse()

	appConfig.stripe.secret = os.Getenv("STRIPE_SECRET_KEY")
	appConfig.stripe.pubKey = os.Getenv("STRIPE_KEY")
	appConfig.db.dsn = os.Getenv("DATABASE_DSN")
	//root:password@tcp(127.0.0.1:3306)/stripe_app

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := driver.OpenDb(appConfig.db.dsn)

	if err != nil {
		errorLog.Fatal(err)
		return
	}
	defer db.Close()

	// Initialize a new session manager and configure the session lifetime.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	tc := make(map[string]*template.Template)

	app := application{
		config:        appConfig,
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: tc,
		version:       version,
		DB:            models.DBModel{DB: db},
		session:       sessionManager,
	}

	err = app.serve()

	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
