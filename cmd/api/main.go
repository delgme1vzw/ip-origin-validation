package main

import (
	"log"

	"github.com/delgme1vzw/ip-origin-validation/internal/datastore"
	"github.com/delgme1vzw/ip-origin-validation/internal/db"
	"github.com/delgme1vzw/ip-origin-validation/internal/env"
)

var appVersion = "v1"

//	@title			Country Whitelister API
//	@description	API to retrieve country whitelist status based on IP

//	@contact.name	delgme1
//	@contact.email	delgme1vzw@gmail.com

// @BasePath	/v1
func main() {
	//Call our load method and set the version which is used throughout the app
	env.LoadEnv()
	appVersion = env.GetString("APP_VERSION", appVersion)

	//Setup configuration details
	//Not worried about error handling for env because we have a default
	cfg := config{
		addr:   env.GetString("APP_ADDRESS", "localhost:8080"),
		apiURL: env.GetString("EXT_API_URL", "localhost:8080"),

		db: dbConfig{
			addr:         env.GetString("DB_ADDRESS", "postgres://user:fakepw@host/dbname?sslmode=disable"), //sslmode disabled on local, should change for prod
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 15),                                               //max number of open conns to db,
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 10),                                               //max number of idle conns to db, <= maxOpen always
			maxIdleTime:  env.GetDuration("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("BE_ENV", "dev"),
	}

	//Create a db instance
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("database connection pool established")

	//Pass the db instance into our storage
	store := datastore.NewStorage(db)
	app := &application{config: cfg, store: store}
	mux := app.mount()

	log.Fatal(app.run(mux))
}
