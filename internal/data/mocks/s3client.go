package mocks

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3Client(handlers ...roundtripperFunc) *s3.Client {
	handler := DefaultS3Handler()
	if len(handlers) > 0 {
		handler = handlers[0]
	}

	return s3.New(s3.Options{
		Credentials: aws.AnonymousCredentials{},
		Region:      "auto",
		HTTPClient:  NewHTTPClient(handler),
	})
}

func DefaultS3Handler() roundtripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		status := http.StatusOK
		if req.Method == http.MethodDelete {
			status = http.StatusNoContent
		}

		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Body:       http.NoBody,
			Header:     make(http.Header),
		}, nil
	}
}
