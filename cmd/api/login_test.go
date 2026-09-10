package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sharasha07/royale-tourneys/internal/assert"
)

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  int
		checkBody func(t *testing.T, resp *http.Response)
	}{
		{
			name:     "Success",
			input:    `{"username": "shaba", "password": "shaba123"}`,
			wantCode: http.StatusOK,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					AccessToken  string `json:"access_token"`
					RefreshToken string `json:"refresh_token"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				if result.AccessToken == "" || result.RefreshToken == "" {
					t.Error("expected non-empty access/refresh tokens")
				}
			},
		},
		{
			name:     "Non-existent user",
			input:    `{"username": "luka", "password": "luka123"}`,
			wantCode: http.StatusUnauthorized,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "invalid credentials", result.Error)
			},
		},
		{
			name:     "Incorrect password",
			input:    `{"username": "shaba", "password": "shaba1234"}`,
			wantCode: http.StatusUnauthorized,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "invalid credentials", result.Error)
			},
		},
	}

	app := newTestApplication()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(tt.input))
			rr := httptest.NewRecorder()

			app.loginHandler(rr, req)
			resp := rr.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			if tt.checkBody != nil {
				tt.checkBody(t, resp)
			}
		})
	}
}

func TestRefreshTokenHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  int
		checkBody func(t *testing.T, resp *http.Response)
	}{
		{
			name:     "Success",
			input:    `{"refresh_token": "token_example"}`,
			wantCode: http.StatusOK,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					AccessToken string `json:"access_token"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				if result.AccessToken == "" {
					t.Error("expected non-empty access token")
				}
			},
		},
		{
			name:     "Empty refresh token",
			input:    `{"refresh_token": ""}`,
			wantCode: http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error map[string]string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "must be provided", result.Error["refresh_token"])
			},
		},
		{
			name:     "Non-existent refresh token",
			input:    `{"refresh_token": "blabla"}`,
			wantCode: http.StatusUnauthorized,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "invalid authentication token", result.Error)
			},
		},
	}

	app := newTestApplication()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(tt.input))
			rr := httptest.NewRecorder()

			app.refreshTokenHandler(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			if tt.checkBody != nil {
				tt.checkBody(t, resp)
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  int
		checkBody func(t *testing.T, resp *http.Response)
	}{
		{
			name:      "Success",
			input:     `{"refresh_token": "token_example"}`,
			wantCode:  http.StatusNoContent,
			checkBody: nil,
		},
		{
			name:     "Empty refresh token",
			input:    `{"refresh_token": ""}`,
			wantCode: http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error map[string]string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "must be provided", result.Error["refresh_token"])
			},
		},
		{
			name:      "Non-existent refresh token",
			input:     `{"refresh_token": "blablabla"}`,
			wantCode:  http.StatusNoContent,
			checkBody: nil,
		},
	}

	app := newTestApplication()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", strings.NewReader(tt.input))
			rr := httptest.NewRecorder()

			app.logoutHandler(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantCode, resp.StatusCode)

			if tt.checkBody != nil {
				tt.checkBody(t, resp)
			}
		})
	}
}
