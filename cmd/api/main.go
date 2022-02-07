package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jersonsatoru/lets-go-further/internal/data"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	version = "1.0.0"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	cfg    *config
	models data.Models
}

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logger)
}

func main() {
	var cfg config
	appPort, _ := strconv.Atoi(os.Getenv("APP_PORT"))
	maxOpenConns, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	maxIdleConns, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))

	flag.IntVar(&cfg.port, "port", appPort, "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("APP_ENV"), "Environment (development-staging-production)")
	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("DSN"), "Data Source Name (DSN)")
	flag.IntVar(&cfg.db.maxOpenConns, "maxOpenConns", maxOpenConns, "Max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "maxIdleConns", maxIdleConns, "Max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "maxIdleTime", os.Getenv("DB_MAX_IDLE_TIME"), "Max idle connections")
	flag.Parse()

	db, err := openDB(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &application{
		cfg:    &cfg,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	zap.L().Info("server is running, with database connection",
		zap.Int("port", cfg.port),
		zap.String("env", cfg.env),
		zap.Bool("db", true))
	err = srv.ListenAndServe()
	zap.L().Fatal("server failed", zap.String("error", err.Error()))
}

func openDB(cfg *config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	d, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(d)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, err
}
