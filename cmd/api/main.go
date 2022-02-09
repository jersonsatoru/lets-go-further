package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
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
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type application struct {
	cfg    *config
	models data.Models
}

type zapLogger struct {
	logger *zap.Logger
}

func (zl *zapLogger) Write(p []byte) (n int, err error) {
	zl.logger.Sugar().Error(string(p))
	return len(p), nil
}

func init() {
	logger, err := zap.NewProduction()
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
	limiterRPS, _ := strconv.ParseFloat(os.Getenv("LIMITER_RPS"), 64)
	limiterBurst, _ := strconv.Atoi(os.Getenv("LIMITER_BURST"))
	limiterEnabled, _ := strconv.ParseBool(os.Getenv("LIMITER_ENABLED"))

	flag.IntVar(&cfg.port, "port", appPort, "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("APP_ENV"), "Environment (development-staging-production)")
	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("DSN"), "Data Source Name (DSN)")
	flag.IntVar(&cfg.db.maxOpenConns, "maxOpenConns", maxOpenConns, "Max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "maxIdleConns", maxIdleConns, "Max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "maxIdleTime", os.Getenv("DB_MAX_IDLE_TIME"), "Max idle connections")

	flag.Float64Var(&cfg.limiter.rps, "limiterRPS", limiterRPS, "RPS Limiter")
	flag.IntVar(&cfg.limiter.burst, "limiterBurst", limiterBurst, "BUrst Limiter")
	flag.BoolVar(&cfg.limiter.enabled, "limiterEnabled", limiterEnabled, "Enable Limiter")
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
	zap.S().Infow("server is running, with database connection",
		"port", cfg.port,
		"env", cfg.env,
		"db", true)
	if err = app.serve(); err != nil {
		zap.S().Fatalw("server failed", zap.String("error", err.Error()))
	}
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
