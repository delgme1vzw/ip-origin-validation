package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// Internal packages shouldn't know about the external world so not passing in context
func New(addr string, maxOpenConns int, maxIdleConns int, maxIdleTime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(maxIdleTime)

	// Create context with timeout of 5 seconds
	// If timeout is not reached we need to cancel (clear) the resources associated with it
	// No timeout means the process was quick enough to connect (to the db in this case)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//setup connection if it isn't up already by using Ping
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil

}
