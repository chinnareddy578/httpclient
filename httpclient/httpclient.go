package httpclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"
)

// WithRetry configures retry count and delay.
func WithRetry(retryCount int, retryDelay time.Duration) Option {
	return func(hc *HTTPClient) {
		hc.retryCount = retryCount
		hc.retryDelay = retryDelay
	}
}

// WithExponentialBackoff configures exponential backoff for retries.
func WithExponentialBackoff(baseDelay time.Duration) Option {
	return func(hc *HTTPClient) {
		hc.backoff = func(attempt int) time.Duration {
			return time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
		}
	}
}

// WithLogger sets a custom logger for the HTTP client.
func WithLogger(logger *log.Logger) Option {
	return func(hc *HTTPClient) {
		hc.logger = logger
	}
}

// WithTransport sets a custom Transport for the HTTP client.
func WithTransport(transport http.RoundTripper) Option {
	return func(hc *HTTPClient) {
		hc.client.Transport = transport
	}
}

// WithTLSConfig sets a custom TLS configuration for the HTTP client.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(hc *HTTPClient) {
		if transport, ok := hc.client.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = tlsConfig
		} else {
			hc.client.Transport = &http.Transport{TLSClientConfig: tlsConfig}
		}
	}
}

// Add default headers to the HTTPClient.
func WithDefaultHeaders(headers map[string]string) Option {
	return func(hc *HTTPClient) {
		if hc.client.Transport == nil {
			hc.client.Transport = &http.Transport{}
		}
		originalTransport := hc.client.Transport
		hc.client.Transport = &headerTransport{
			base:    originalTransport,
			headers: headers,
		}
	}
}

// WithTimeout configures the timeout for the HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(hc *HTTPClient) {
		hc.client.Timeout = timeout
	}
}

// NewHTTPClient creates a new instance of HTTPClient with the provided options.
func NewHTTPClient(options ...Option) *HTTPClient {
	hc := &HTTPClient{
		client:  &http.Client{},
		logger:  log.Default(),
		backoff: func(attempt int) time.Duration { return 0 }, // Default: no backoff
	}
	for _, opt := range options {
		opt(hc)
	}
	return hc
}

// Do sends an HTTP request and returns an HTTP response, with retry and logging.
func (hc *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error
	for i := 0; i <= hc.retryCount; i++ {
		if i > 0 {
			delay := hc.retryDelay
			if hc.backoff != nil {
				delay = hc.backoff(i)
			}
			hc.logger.Printf("Retrying request (%d/%d) after %v...", i, hc.retryCount, delay)
			time.Sleep(delay)
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

// Put is a helper method for making PUT requests.
func (hc *HTTPClient) Put(url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return hc.Do(req)
}

// Delete is a helper method for making DELETE requests.
func (hc *HTTPClient) Delete(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return hc.Do(req)
}

// PostJSON is a helper method for making POST requests with JSON body.
func (hc *HTTPClient) PostJSON(url string, jsonBody interface{}, headers map[string]string) (*http.Response, error) {
	// Ensure headers map is initialized before adding Content-Type.
	if headers == nil {
		headers = make(map[string]string)
	}
	body, err := json.Marshal(jsonBody)
	if err != nil {
		return nil, err
	}
	headers["Content-Type"] = "application/json"
	return hc.Post(url, bytes.NewReader(body), headers)
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

// ReadJSONResponseBody reads and unmarshals the response body into the target interface.
func ReadJSONResponseBody(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
