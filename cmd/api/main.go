package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type config struct {
	port int
}

func main() {
	godotenv.Load()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Fatal(srv.ListenAndServe())
}

func loadConfig() (*config, error) {
	var cfg config
	if port, ok := os.LookupEnv("PORT"); ok {
		portNumber, err := strconv.Atoi(port)
		if err != nil {
			return nil, errors.New("PORT must be number")
		}
		cfg.port = portNumber
	} else {
		return nil, errors.New("PORT must be provided")
	}

	return &cfg, nil
}
