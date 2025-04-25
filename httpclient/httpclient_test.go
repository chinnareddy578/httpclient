package httpclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	logger := log.Default()
	client := NewHTTPClient(
		WithTimeout(5*time.Second),
		WithRetry(3, 1*time.Second),
		WithLogger(logger),
	)

	if client.client.Timeout != 5*time.Second {
		t.Errorf("Expected timeout to be 5 seconds, got %v", client.client.Timeout)
	}
	if client.retryCount != 3 {
		t.Errorf("Expected retry count to be 3, got %d", client.retryCount)
	}
	if client.retryDelay != 1*time.Second {
		t.Errorf("Expected retry delay to be 1 second, got %v", client.retryDelay)
	}
	if client.logger != logger {
		t.Errorf("Expected logger to be set correctly")
	}
}

func TestNewHTTPClient_WithExponentialBackoff(t *testing.T) {
	client := NewHTTPClient(
		WithRetry(3, 0),
		WithExponentialBackoff(100*time.Millisecond),
	)

	expectedDelays := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
	}

	for i := 1; i <= 3; i++ {
		delay := client.backoff(i)
		if delay != expectedDelays[i-1] {
			t.Errorf("Expected delay for attempt %d to be %v, got %v", i, expectedDelays[i-1], delay)
		}
	}
}

func TestDo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	logger := log.Default()
	client := NewHTTPClient(
		WithTimeout(5*time.Second),
		WithRetry(0, 0),
		WithLogger(logger),
	)

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestDo_RetryWithExponentialBackoff(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			http.Error(w, "Temporary error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	logger := log.Default()
	client := NewHTTPClient(
		WithRetry(3, 0),
		WithExponentialBackoff(100*time.Millisecond),
		WithLogger(logger),
	)

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	logger := log.Default()
	client := NewHTTPClient(
		WithTimeout(5*time.Second),
		WithRetry(0, 0),
		WithLogger(logger),
	)

	headers := map[string]string{
		"Authorization": "Bearer token",
	}
	resp, err := client.Get(server.URL, headers)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "test payload") {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	logger := log.Default()
	client := NewHTTPClient(
		WithTimeout(5*time.Second),
		WithRetry(0, 0),
		WithLogger(logger),
	)

	body := bytes.NewBufferString("test payload")
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := client.Post(server.URL, body, headers)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestReadResponseBody(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader("response body")),
	}

	body, err := ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if body != "response body" {
		t.Errorf("Expected body to be 'response body', got '%s'", body)
	}
}

func TestReadResponseBody_Error(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(&errorReader{}),
	}

	_, err := ReadResponseBody(resp)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

func TestWithTransport(t *testing.T) {
	customTransport := &http.Transport{}
	client := NewHTTPClient(WithTransport(customTransport))

	if client.client.Transport != customTransport {
		t.Errorf("Expected custom transport to be set")
	}
}

func TestWithTLSConfig(t *testing.T) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	client := NewHTTPClient(WithTLSConfig(tlsConfig))

	transport, ok := client.client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Expected transport to be of type *http.Transport")
	}
	if transport.TLSClientConfig != tlsConfig {
		t.Errorf("Expected TLS config to be set")
	}
}

func TestHTTPClient_Get(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET response"))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	hc := NewHTTPClient()
	resp, err := hc.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestHTTPClient_PostJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("POST response"))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	hc := NewHTTPClient()
	resp, err := hc.PostJSON(ts.URL, map[string]string{"key": "value"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestHTTPClient_Retry(t *testing.T) {
	attempts := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	hc := NewHTTPClient(WithRetry(3, time.Millisecond))
	resp, err := hc.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestHTTPClient_DefaultHeaders(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "value" {
			t.Errorf("expected X-Custom-Header value, got %s", r.Header.Get("X-Custom-Header"))
		}
		w.WriteHeader(http.StatusOK)
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	hc := NewHTTPClient(WithDefaultHeaders(map[string]string{"X-Custom-Header": "value"}))
	resp, err := hc.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
