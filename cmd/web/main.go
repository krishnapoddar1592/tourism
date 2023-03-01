package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Atul-Ranjan12/tourism/internal/config"
	"github.com/Atul-Ranjan12/tourism/internal/handlers"
	"github.com/Atul-Ranjan12/tourism/internal/helpers"
	"github.com/Atul-Ranjan12/tourism/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":4000"

var app config.AppConfig
var session *scs.SessionManager

func main() {
	err := run()
	if err != nil {
		log.Println("Error executing the run function")
	}
	// TODO: Create and handle mailing channel

	fmt.Println("Starting Server on localhost port", portNumber)

	// Set up server configurations:
	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	serverError := server.ListenAndServe()
	log.Fatal(serverError)
}

// run handles major initialization processes for the app
func run() error {
	// Set up configuration variable to not be in production mode
	app.InProduction = false

	// Set up Error and information log
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime)

	// Initialize the session for the application
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	// Add the session to the application config
	app.Session = session

	// TODO: Handle connection to the database in this section

	// Handle creation of template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return err
	}
	app.TemplateCache = tc
	app.UseCache = true

	// Create Handlers and Repository
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	// Create Renderer and Helpers
	render.NewRenderer(&app)
	helpers.NewHelper(&app)

	return nil
}
