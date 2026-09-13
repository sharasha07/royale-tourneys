package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
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
			R2: struct {
				AccessKey       string `env:"R2_ACCESS_KEY,required"`
				SecretAccessKey string `env:"R2_SECRET_ACCESS_KEY,required"`
				Bucket          string `env:"R2_BUCKET,required"`
				PublicURL       string `env:"R2_PUBLIC_URL,required"`
				S3ApiEndpoint   string `env:"S3_API_ENDPOINT,required"`
			}{
				Bucket:    "test-bucket",
				PublicURL: "https://cdn.example.com",
			},
		},
		models: data.Models{
			Users:  mocks.NewUserModel(),
			Tokens: mocks.NewTokenModel(),
		},
		httpClient: mocks.NewHTTPClient(mocks.StatusRouter(nil)),
		s3Client:   mocks.NewS3Client(),
	}
}

func makeImage(t *testing.T, format string) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	var buf bytes.Buffer
	switch format {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
	case "jpeg":
		if err := jpeg.Encode(&buf, img, nil); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unsupported format %q", format)
	}
	return buf.Bytes()
}
