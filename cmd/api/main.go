package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

type config struct {
	port          int
	postgresURL   string
	clashAPIToken string
}

type application struct {
	cfg   config
	model data.DBModel
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.postgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to the database")

	app := &application{
		cfg:   cfg,
		model: data.NewDBModel(pool),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	mux.HandleFunc("POST /v1/users", app.createUserHandler)
	mux.HandleFunc("GET /v1/users/{id}", app.showUserHandler)
	mux.HandleFunc("PATCH /v1/users/{id}", app.updateUserHandler)
	mux.HandleFunc("UPDATE /v1/users/{id}/tag", app.updateGameTagHandler)
	mux.HandleFunc("DELETE /v1/users/{id}", app.deleteUserHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      recoverPanic(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Printf("starting server on port: %d", cfg.port)
	err = srv.ListenAndServe()
	log.Fatal(err)
}

func loadConfig() (config, error) {
	var cfg config
	if port, ok := os.LookupEnv("PORT"); ok {
		portNumber, err := strconv.Atoi(port)
		if err != nil {
			return config{}, errors.New("PORT must be number")
		}
		cfg.port = portNumber
	} else {
		return config{}, errors.New("PORT must be provided")
	}

	if postgresURL, ok := os.LookupEnv("POSTGRES_URL"); !ok {
		return config{}, errors.New("POSTGRES_URL must be provided")
	} else {
		cfg.postgresURL = postgresURL
	}

	if token, ok := os.LookupEnv("CLASH_API_TOKEN"); !ok {
		return config{}, errors.New("CLASH_API_TOKEN must be provided")
	} else {
		cfg.clashAPIToken = token
	}

	return cfg, nil
}
