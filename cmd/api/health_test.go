package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sharasha07/royale-tourneys/internal/assert"
)

func TestHealthHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	healthHandler(rr, r)

	rs := rr.Result()
	defer rs.Body.Close()

	assert.Equal(t, rs.StatusCode, http.StatusOK)
	assert.Equal(t, rs.Header.Get("Content-Type"), "application/json")

	var got struct {
		Status string `json:"status"`
	}

	err := json.NewDecoder(rs.Body).Decode(&got)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, got.Status, "available")
}
