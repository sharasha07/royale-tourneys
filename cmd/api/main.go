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

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

type config struct {
	port              int
	postgresURL       string
	jwtSecret         string
	clashAPIToken     string
	r2AccountID       string
	r2AccessKey       string
	r2SecretAccessKey string
	r2Bucket          string
	r2PublicURL       string
	s3ApiEndpoint     string
}

type application struct {
	cfg      config
	model    data.DBModel
	client   *http.Client
	s3Client *s3.Client
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

	s3Client := s3.New(s3.Options{
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.r2AccessKey, cfg.r2SecretAccessKey, ""),
		Region:       "auto",
		BaseEndpoint: &cfg.s3ApiEndpoint,
		UsePathStyle: true,
	})

	app := &application{
		cfg:      cfg,
		model:    data.NewDBModel(pool),
		client:   &http.Client{Timeout: 10 * time.Second},
		s3Client: s3Client,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	mux.HandleFunc("POST /v1/users", app.createUserHandler)
	mux.HandleFunc("GET /v1/users/{id}", app.showUserHandler)
	mux.HandleFunc("PATCH /v1/users/{id}", app.updateUserHandler)
	mux.HandleFunc("PUT /v1/users/{id}/tag", app.updateGameTagHandler)
	mux.HandleFunc("PUT /v1/users/{id}/profile_picture", app.updateProfilePictureHandler)
	mux.HandleFunc("DELETE /v1/users/{id}", app.deleteUserHandler)

	mux.HandleFunc("POST /v1/tokens", app.createAuthenticationTokenHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      recoverPanic(app.authenticate(mux)),
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

	if jwtSecret, ok := os.LookupEnv("JWT_SECRET"); !ok {
		return config{}, errors.New("JWT_SECRET must be provided")
	} else {
		cfg.jwtSecret = jwtSecret
	}

	if token, ok := os.LookupEnv("CLASH_API_TOKEN"); !ok {
		return config{}, errors.New("CLASH_API_TOKEN must be provided")
	} else {
		cfg.clashAPIToken = token
	}

	if r2AccountID, ok := os.LookupEnv("R2_ACCOUNT_ID"); !ok {
		return config{}, errors.New("R2_ACCOUNT_ID must be provided")
	} else {
		cfg.r2AccountID = r2AccountID
	}

	if r2AccessKey, ok := os.LookupEnv("R2_ACCESS_KEY"); !ok {
		return config{}, errors.New("R2_ACCESS_KEY must be provided")
	} else {
		cfg.r2AccessKey = r2AccessKey
	}

	if r2SecretAccessKey, ok := os.LookupEnv("R2_SECRET_ACCESS_KEY"); !ok {
		return config{}, errors.New("R2_SECRET_ACCESS_KEY must be provided")
	} else {
		cfg.r2SecretAccessKey = r2SecretAccessKey
	}

	if r2Bucket, ok := os.LookupEnv("R2_BUCKET"); !ok {
		return config{}, errors.New("R2_BUCKET must be provided")
	} else {
		cfg.r2Bucket = r2Bucket
	}

	if r2PublicURL, ok := os.LookupEnv("R2_PUBLIC_URL"); !ok {
		return config{}, errors.New("R2_PUBLIC_URL must be provided")
	} else {
		cfg.r2PublicURL = r2PublicURL
	}

	if s3APIEndpoint, ok := os.LookupEnv("S3_API_ENDPOINT"); !ok {
		return config{}, errors.New("S3_API_ENDPOINT must be provided")
	} else {
		cfg.s3ApiEndpoint = s3APIEndpoint
	}

	return cfg, nil
}
