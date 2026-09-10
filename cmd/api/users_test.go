package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/sharasha07/royale-tourneys/internal/assert"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  int
		checkBody func(t *testing.T, resp *http.Response)
	}{
		{
			name:     "Success",
			input:    `{"username": "saba", "password": "saba123"}`,
			wantCode: http.StatusCreated,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()
				var result struct {
					User data.User `json:"user"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "saba", result.User.Username)
				assert.Equal(t, time.Now().Year(), result.User.CreatedAt.Year())
				assert.Equal(t, time.Now().Weekday(), result.User.CreatedAt.Weekday())
			},
		},
		{
			name:     "Username Unique Violation",
			input:    `{"username": "shaba", "password": "shaba123"}`,
			wantCode: http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()
				var result struct {
					Error map[string]string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "must be unique", result.Error["username"])
			},
		},
		{
			name:     "Invalid Username",
			input:    `{"username": "shabashabashaba", "password": "shaba123"}`,
			wantCode: http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()
				var result struct {
					Error map[string]string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "must not have more than 10 characters", result.Error["username"])
			},
		},
		{
			name:     "Invalid Password",
			input:    `{"username": "shaba", "password": "sha"}`,
			wantCode: http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()
				var result struct {
					Error map[string]string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "must be more than 5 characters", result.Error["password"])
			},
		},
		{
			name:     "Invalid Username and Password",
			input:    `{"username": "shabashabashaba", "password": "sha"}`,
			wantCode: http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()
				var result struct {
					Error map[string]string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "must not have more than 10 characters", result.Error["username"])
				assert.Equal(t, "must be more than 5 characters", result.Error["password"])
			},
		},
	}

	app := newTestApplication()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/users", strings.NewReader(tt.input))
			rr := httptest.NewRecorder()

			app.createUserHandler(rr, req)

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

func TestShowUserHandler(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantCode  int
		checkBody func(t *testing.T, resp *http.Response)
	}{
		{
			name:     "SUCCESS",
			id:       "1",
			wantCode: http.StatusOK,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()

				var result struct {
					User data.User `json:"user"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, 1, result.User.ID)
				assert.Equal(t, "shaba", result.User.Username)
				assert.Equal(t, 2, result.User.Version)
			},
		},
		{
			name:     "Decimal ID",
			id:       "1.5",
			wantCode: http.StatusNotFound,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()

				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "resource not found", result.Error)
			},
		},
		{
			name:     "Negative ID",
			id:       "-10",
			wantCode: http.StatusNotFound,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()

				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "resource not found", result.Error)
			},
		},
		{
			name:     "Non-existent ID",
			id:       "15",
			wantCode: http.StatusNotFound,
			checkBody: func(t *testing.T, resp *http.Response) {
				t.Helper()

				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "resource not found", result.Error)
			},
		},
	}

	app := newTestApplication()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/users/"+tt.id, nil)
			rr := httptest.NewRecorder()
			req.SetPathValue("id", tt.id)

			app.showUserHandler(rr, req)

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

func TestUpdateUserHandler(t *testing.T) {
	hash, err := argon2id.CreateHash("shaba123", argon2id.DefaultParams)
	if err != nil {
		t.Fatal(err)
	}

	u := &data.User{
		ID:           1,
		Username:     "shaba",
		PasswordHash: []byte(hash),
		CreatedAt:    time.Now(),
		Version:      2,
	}

	tests := []struct {
		name      string
		id        string
		input     string
		user      *data.User
		wantCode  int
		checkBody func(t *testing.T, resp *http.Response)
	}{
		{
			name:     "Anonymous user",
			id:       "1",
			input:    `{"username": "baba", "password" "bubu"}`,
			user:     data.AnonymousUser,
			wantCode: http.StatusUnauthorized,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "authorization required to access this endpoint", result.Error)
			},
		},
		{
			name:     "Not permitted, valid user ID",
			id:       "15",
			input:    `{"username": "baba", "password" "bubu"}`,
			user:     u,
			wantCode: http.StatusForbidden,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "user is not permitted to access this resource", result.Error)
			},
		},
		{
			name:     "permitted, valid user ID, valid input",
			id:       "1",
			user:     u,
			input:    `{"username": "babi", "password": "babi123"}`,
			wantCode: http.StatusOK,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					User data.User `json:"user"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, 1, result.User.ID)
				assert.Equal(t, "babi", result.User.Username)
			},
		},
		{
			name:     "permitted, decimal user ID, valid input",
			id:       "1.5",
			user:     u,
			input:    `{"username": "babi" "password": "babi123"}`,
			wantCode: http.StatusNotFound,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "resource not found", result.Error)
			},
		},
		{
			name:     "permitted, negative user ID, valid input",
			id:       "-15",
			user:     u,
			input:    `{"username": "babi" "password": "babi123"}`,
			wantCode: http.StatusNotFound,
			checkBody: func(t *testing.T, resp *http.Response) {
				var result struct {
					Error string `json:"error"`
				}

				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, "resource not found", result.Error)
			},
		},
	}

	app := newTestApplication()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/v1/users/"+tt.id, strings.NewReader(tt.input))
			req.SetPathValue("id", tt.id)
			req = contextSetUser(req, tt.user)
			rr := httptest.NewRecorder()

			app.updateUserHandler(rr, req)
			resp := rr.Result()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			if tt.checkBody != nil {
				tt.checkBody(t, resp)
			}
		})
	}
}
