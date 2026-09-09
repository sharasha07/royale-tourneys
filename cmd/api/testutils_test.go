package main

import (
	"time"

	"github.com/sharasha07/royale-tourneys/internal/data"
	"github.com/sharasha07/royale-tourneys/internal/data/mocks"
)

func newTestApplication() *application {
	return &application{
		cfg: Config{
			JWT: struct {
				Secret     string        `env:"JWT_SECRET,required"`
				AccessTTL  time.Duration `env:"JWT_ACCESS_TTL,required"`
				RefreshTTL time.Duration `env:"JWT_REFRESH_TTL,required"`
			}{
				Secret:     "this-is-test-secret-this-is-test-secret-this-is-test-secret",
				AccessTTL:  15 * time.Minute,
				RefreshTTL: 24 * time.Hour,
			},
		},
		models: data.Models{
			Users:  mocks.NewUserModel(),
			Tokens: mocks.NewTokenModel(),
		},
	}
}
