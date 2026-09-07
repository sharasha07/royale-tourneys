package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sharasha07/royale-tourneys/internal/assert"
)

func TestHealthHandler(t *testing.T) {
	app := &application{}

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	assert.Equal(t, rs.StatusCode, http.StatusOK)
	assert.Equal(t, rs.Header.Get("Content-Type"), "application/json")

	var got struct {
		Status string `json:"status"`
	}

	err = json.NewDecoder(rs.Body).Decode(&got)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, got.Status, "available")
}
