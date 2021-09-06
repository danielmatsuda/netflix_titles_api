package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"danielmatsuda15.rest/internal/data"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps   float64
		burst int
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	var cfg config

	// uses -port and -env flags as config values, defaulting to port 4000 and "development"
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// NOTE: for local testing, the database DSN is automatically provided in the Makefile via environment variable
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	// get connection pool flags
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// rate limiter settings. Rate limiting is always enabled
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter max requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter max burst")

	flag.Parse()

	// init a logger that writes to stdout, prefixed with current date and time
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// create a connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Printf("db connection pool established")

	// publish variables in the expvar handler. View them and other metrics at /debug/vars
	expvar.NewString("version").Set(version)
	// # of active goroutines
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	// db connection pool info
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	// current timestamp
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// new application instance
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db), // create a new Model for the db connection pool
	}

	// declare http server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// openDB returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	// use DSN to create empty conn pool
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// set the settings for the connection pool based on cfg attributes
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	// create a context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// establish new connection using PingContext() and the context with timeout
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
