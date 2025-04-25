package httpclient

import (
	"log"
	"net/http"
	"time"
)

// HTTPClient is a wrapper around the standard http.Client with additional features.
type HTTPClient struct {
	client     *http.Client
	retryCount int
	retryDelay time.Duration
	logger     *log.Logger
	backoff    func(attempt int) time.Duration
}

// Option is a functional option for configuring the HTTPClient.
type Option func(*HTTPClient)

// headerTransport is a custom RoundTripper to add default headers.
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (ht *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range ht.headers {
		req.Header.Set(key, value)
	}
	return ht.base.RoundTrip(req)
}
