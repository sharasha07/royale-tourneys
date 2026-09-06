package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

type Config struct {
	Port          int    `env:"PORT,required"`
	ClashAPIToken string `env:"CLASH_API_TOKEN,required"`

	DB struct {
		DSN         string        `env:"DB_DSN,required"`
		MinConns    int32         `env:"DB_MIN_CONNS,required"`
		MaxConns    int32         `env:"DB_MAX_CONNS,required"`
		MaxIdleTime time.Duration `env:"DB_MAX_IDLE_TIME,required"`
	}

	JWT struct {
		Secret     string        `env:"JWT_SECRET,required"`
		AccessTTL  time.Duration `env:"JWT_ACCESS_TTL,required"`
		RefreshTTL time.Duration `env:"JWT_REFRESH_TTL,required"`
	}

	R2 struct {
		AccessKey       string `env:"R2_ACCESS_KEY,required"`
		SecretAccessKey string `env:"R2_SECRET_ACCESS_KEY,required"`
		Bucket          string `env:"R2_BUCKET,required"`
		PublicURL       string `env:"R2_PUBLIC_URL,required"`
		S3ApiEndpoint   string `env:"S3_API_ENDPOINT,required"`
	}

	Limiter struct {
		RPS     float64 `env:"LIMITER_RPS,required"`
		Burst   int     `env:"LIMITER_BURST,required"`
		Enabled bool    `env:"LIMITER_ENABLED,required"`
	}

	CORS struct {
		TrustedOrigins []string `env:"TRUSTED_ORIGINS,required"`
	}
}

type application struct {
	cfg        Config
	model      data.DBModel
	httpClient *http.Client
	s3Client   *s3.Client
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	pool, err := connectToDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Println("connected to the database")

	s3Client := s3.New(s3.Options{
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.R2.AccessKey, cfg.R2.SecretAccessKey, ""),
		Region:       "auto",
		BaseEndpoint: aws.String(cfg.R2.S3ApiEndpoint),
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

func connectToDB(cfg Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	poolConfig.MinConns = cfg.DB.MinConns
	poolConfig.MaxConns = cfg.DB.MaxConns
	poolConfig.MaxConnIdleTime = cfg.DB.MaxIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
