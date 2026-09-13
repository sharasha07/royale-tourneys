package mocks

import (
	"net/http"
)

type roundtripperFunc func(*http.Request) (*http.Response, error)

func (f roundtripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func NewHTTPClient(handler roundtripperFunc) *http.Client {
	return &http.Client{
		Transport: handler,
	}
}

func StatusRouter(statusByPath map[string]int) roundtripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		status, ok := statusByPath[req.URL.Path]
		if !ok {
			status = http.StatusInternalServerError
		}
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Body:       http.NoBody,
			Header:     make(http.Header),
		}, nil
	}
}
