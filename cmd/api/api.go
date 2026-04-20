package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/delgme1vzw/ip-origin-validation/internal/datastore"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/delgme1vzw/ip-origin-validation/docs" //This is required to generate swagger docs
	httpSwagger "github.com/swaggo/http-swagger"
)

// Create a struct for our api application and config
// so it is cleaner to hold different configuration details
// per environment
// Obviously this is overkill for a POC but good for future reference
type application struct {
	config config
	store  datastore.Storage
}

type config struct {
	addr   string
	db     dbConfig
	env    string
	apiURL string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

// Create a mount method for our application abstraction
// Will be used to track routing when we run() the app
// the return parameter is *chi.mux which implements the http handler,
// so we will use http handler so we do not limit ourselves in functionality
// this allows us to change the handler type in the future if we want
func (app *application) mount() http.Handler {
	// We can use server mux, but chi
	// is built ontop of context and allows
	// the use of middleware and easier handling
	// of routing (inc. version control), allowing
	// us to expand the project if needed
	r := chi.NewRouter()

	// Some good middlewares built in with chi
	// request id generator and RemoteAddr header real ip injection
	// logging and recover from panic
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/"+appVersion, func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		//setting up the docsURL so when we go to (http://localhost:8080/v1/swagger/index.html) it accesses the json
		//Cmd to generate docs: swag init -g ./api/main.go -d cmd,internal  #(inc internal due to datastore obj access)
		docsURL := fmt.Sprintf("/%s/swagger/doc.json", appVersion)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		//I debated rooting this in /ip-map, but the user won't want an ip or country code back,
		//so I feel it makes more sense calling it what the user would request
		r.Route("/whitelist-status", func(r chi.Router) {
			r.Route("/{ip}", func(r chi.Router) {
				r.Get("/", app.getWhitelistStatusHandler)
			})
		})

		//On the other hand, the ip-map root can be used to update the existing map/maintain it

		//POST /v1/ip-map
		//This is strictly for maintaining the routing
		r.Route("/ip-map", func(r chi.Router) {
			//Get the current country code
			//accept an ip
			r.Route("/{ip}", func(r chi.Router) {
				r.Get("/", app.getCountryByIPHandler)
			})

			r.Post("/", app.getCreateMappedCountryHandler)
			r.Put("/", app.getUpdateMappedCountryHandler)
			r.Patch("/", app.getUpdateMappedCountryHandler)
			r.Delete("/", app.getDeleteMappedCountryHandler)
		})

	})

	return r
}

// Create a run method for our application abstraction
func (app *application) run(mux http.Handler) error {

	//Docs
	docs.SwaggerInfo.Version = appVersion
	//docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = ""

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server  has started at %s", app.config.addr)
	return srv.ListenAndServe()
}
