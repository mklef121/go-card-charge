package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mklef121/go-card-charge/driver"
	"github.com/mklef121/go-card-charge/internal/models"
)

const version = "1.0.0"

type Config struct {
	// app port
	port int
	//wether it is production or development
	env string
	// Database connection details
	db struct {
		dsn string
	}

	// Stripe account details
	stripe struct {
		secret string
		pubKey string
	}

	//email config
	smtp struct {
		host     string
		port     int
		username string
		password string
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
	version  string
	DB       models.DBModel
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
		Handler:           app.loadApiRoutes(),
	}

	app.infoLog.Printf("Starting Backend server in %s mode on %d port \n", app.config.env, app.config.port)

	return serverConfig.ListenAndServe()
}

func main() {

	var appConfig Config
	defaultPort := 4001
	flag.IntVar(&appConfig.port, "port", defaultPort, "Server port to listen")
	flag.StringVar(&appConfig.env, "env", "development", "application development environment {development | production}")

	flag.Parse()

	appConfig.stripe.secret = os.Getenv("STRIPE_SECRET_KEY")
	appConfig.stripe.pubKey = os.Getenv("STRIPE_KEY")
	appConfig.db.dsn = os.Getenv("DATABASE_DSN")

	appConfig.smtp.host = os.Getenv("MAIL_HOST")
	appConfig.smtp.port, _ = strconv.Atoi(os.Getenv("MAIL_PORT"))
	appConfig.smtp.username = os.Getenv("MAIL_USERNAME")
	appConfig.smtp.password = os.Getenv("MAIL_PASSWORD")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := driver.OpenDb(appConfig.db.dsn)

	if err != nil {
		errorLog.Fatal(err)
		return
	}
	defer db.Close()

	app := application{
		config:   appConfig,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       models.DBModel{DB: db},
	}

	err = app.serve()

	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
