package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

type config struct {
	port        int
	postgresURL string

	jwt struct {
		secret     string
		accessTTL  time.Duration
		refreshTTL time.Duration
	}
	clashAPIToken string

	r2 struct {
		accessKey       string
		secretAccessKey string
		bucket          string
		publicURL       string
		s3ApiEndpoint   string
	}

	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	cfg        config
	model      data.DBModel
	httpClient *http.Client
	s3Client   *s3.Client
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
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.r2.accessKey, cfg.r2.secretAccessKey, ""),
		Region:       "auto",
		BaseEndpoint: aws.String(cfg.r2.s3ApiEndpoint),
		UsePathStyle: true,
	})

	app := &application{
		cfg:        cfg,
		model:      data.NewDBModel(pool),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		s3Client:   s3Client,
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

func loadConfig() (config, error) {
	var cfg config

	if port, ok := os.LookupEnv("PORT"); !ok {
		return config{}, errors.New("PORT environment variable must be provided")
	} else {
		portNumber, err := strconv.Atoi(port)
		if err != nil {
			return config{}, errors.New("PORT must be a number")
		}
		cfg.port = portNumber
	}

	if postgresURL, ok := os.LookupEnv("POSTGRES_URL"); !ok {
		return config{}, errors.New("POSTGRES_URL environment variable must be provided")
	} else {
		cfg.postgresURL = postgresURL
	}

	if jwtSecret, ok := os.LookupEnv("JWT_SECRET"); !ok {
		return config{}, errors.New("JWT_SECRET environment variable must be provided")
	} else {
		cfg.jwt.secret = jwtSecret
	}

	if jwtAccessTTL, ok := os.LookupEnv("JWT_ACCESS_TTL"); !ok {
		return config{}, errors.New("JWT_ACCESS_TTL environment variable must be provided")
	} else {
		dur, err := time.ParseDuration(jwtAccessTTL)
		if err != nil {
			return config{}, err
		}

		cfg.jwt.accessTTL = dur
	}

	if jwtRefreshTTL, ok := os.LookupEnv("JWT_REFRESH_TTL"); !ok {
		return config{}, errors.New("JWT_REFRESH_TTL environment variable must be provided")
	} else {
		dur, err := time.ParseDuration(jwtRefreshTTL)
		if err != nil {
			return config{}, err
		}

		cfg.jwt.refreshTTL = dur
	}

	if token, ok := os.LookupEnv("CLASH_API_TOKEN"); !ok {
		return config{}, errors.New("CLASH_API_TOKEN environment variable must be provided")
	} else {
		cfg.clashAPIToken = token
	}

	if r2AccessKey, ok := os.LookupEnv("R2_ACCESS_KEY"); !ok {
		return config{}, errors.New("R2_ACCESS_KEY environment variable must be provided")
	} else {
		cfg.r2.accessKey = r2AccessKey
	}

	if r2SecretAccessKey, ok := os.LookupEnv("R2_SECRET_ACCESS_KEY"); !ok {
		return config{}, errors.New("R2_SECRET_ACCESS_KEY environment variable must be provided")
	} else {
		cfg.r2.secretAccessKey = r2SecretAccessKey
	}

	if r2Bucket, ok := os.LookupEnv("R2_BUCKET"); !ok {
		return config{}, errors.New("R2_BUCKET environment variable must be provided")
	} else {
		cfg.r2.bucket = r2Bucket
	}

	if r2PublicURL, ok := os.LookupEnv("R2_PUBLIC_URL"); !ok {
		return config{}, errors.New("R2_PUBLIC_URL environment variablemust be provided")
	} else {
		cfg.r2.publicURL = r2PublicURL
	}

	if s3APIEndpoint, ok := os.LookupEnv("S3_API_ENDPOINT"); !ok {
		return config{}, errors.New("S3_API_ENDPOINT environment variable must be provided")
	} else {
		cfg.r2.s3ApiEndpoint = s3APIEndpoint
	}

	if limiterRPS, ok := os.LookupEnv("LIMITER_RPS"); !ok {
		return config{}, errors.New("LIMITER_RPS environment variable must be provided")
	} else {
		rpsNumber, err := strconv.ParseFloat(limiterRPS, 64)
		if err != nil {
			return config{}, errors.New("LIMITER_RPS must be a float")
		}
		cfg.limiter.rps = rpsNumber
	}

	if limiterBurst, ok := os.LookupEnv("LIMITER_BURST"); !ok {
		return config{}, errors.New("LIMITER_BURST environment variable must be provided")
	} else {
		burstNumber, err := strconv.Atoi(limiterBurst)
		if err != nil {
			return config{}, errors.New("LIMITER_RPS must be an int")
		}
		cfg.limiter.burst = burstNumber
	}

	if limiterEnabled, ok := os.LookupEnv("LIMITER_ENABLED"); !ok {
		return config{}, errors.New("LIMITER_ENABLED environment variable must be provided")
	} else {
		ok, err := strconv.ParseBool(limiterEnabled)
		if err != nil {
			return config{}, errors.New("LIMITER_ENABLED must be a bool")
		}
		cfg.limiter.enabled = ok
	}

	if trustedOrigins, ok := os.LookupEnv("TRUSTED_ORIGINS"); !ok {
		return config{}, errors.New("TRUSTED_ORIGINS environment variable must be provided")
	} else {
		cfg.cors.trustedOrigins = strings.Split(trustedOrigins, ",")
	}

	return cfg, nil
}
