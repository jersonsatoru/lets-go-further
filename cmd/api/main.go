package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	version = "1.0.0"
)

type config struct {
	port int
	env  string
}

type application struct {
	cfg *config
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
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development-staging-production)")
	flag.Parse()

	app := &application{
		cfg: &cfg,
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	zap.L().Info("server is running", zap.Int("port", cfg.port), zap.String("env", cfg.env))
	err := srv.ListenAndServe()
	zap.L().Fatal("server failed", zap.String("error", err.Error()))
}
