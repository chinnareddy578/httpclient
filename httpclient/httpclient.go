package httpclient

import (
	"errors"
	"io"
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
}

// NewHTTPClient creates a new instance of HTTPClient.
func NewHTTPClient(timeout time.Duration, retryCount int, retryDelay time.Duration, logger *log.Logger) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		retryCount: retryCount,
		retryDelay: retryDelay,
		logger:     logger,
	}
}

// Do sends an HTTP request and returns an HTTP response, with retry and logging.
func (hc *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error
	for i := 0; i <= hc.retryCount; i++ {
		if i > 0 {
			hc.logger.Printf("Retrying request (%d/%d)...", i, hc.retryCount)
			time.Sleep(hc.retryDelay)
		}

		resp, err := hc.client.Do(req)
		if err != nil {
			hc.logger.Printf("Request failed: %v", err)
			lastErr = err
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		hc.logger.Printf("Received non-2xx response: %d", resp.StatusCode)
		lastErr = errors.New("non-2xx response received")
		resp.Body.Close()
	}

	return nil, lastErr
}

// Get is a helper method for making GET requests.
func (hc *HTTPClient) Get(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return hc.Do(req)
}

// Post is a helper method for making POST requests.
func (hc *HTTPClient) Post(url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return hc.Do(req)
}

// ReadResponseBody reads and returns the response body as a string.
func ReadResponseBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
