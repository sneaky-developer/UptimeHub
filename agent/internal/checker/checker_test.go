package checker

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckHTTPUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewChecker(5 * time.Second)
	result := c.check(context.Background(), Target{ServiceID: "svc-1", Type: "http", URL: srv.URL})

	if !result.IsUp {
		t.Fatalf("expected service up, got down: %s", result.ErrorMessage)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestCheckHTTPDownOnServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewChecker(5 * time.Second)
	result := c.check(context.Background(), Target{ServiceID: "svc-1", Type: "http", URL: srv.URL})

	if result.IsUp {
		t.Fatal("expected service down on 500, got up")
	}
	if result.ErrorMessage == "" {
		t.Error("expected error message for unhealthy status code")
	}
}

func TestCheckHTTPDownOnTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer srv.Close()

	c := NewChecker(5 * time.Second)
	result := c.check(context.Background(), Target{
		ServiceID: "svc-1",
		Type:      "http",
		URL:       srv.URL,
		Timeout:   50 * time.Millisecond,
	})

	if result.IsUp {
		t.Fatal("expected service down on timeout, got up")
	}
}

func TestCheckHTTPDownOnConnectionRefused(t *testing.T) {
	// Grab a port that nothing is listening on
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	url := "http://" + l.Addr().String()
	l.Close()

	c := NewChecker(2 * time.Second)
	result := c.check(context.Background(), Target{ServiceID: "svc-1", Type: "http", URL: url})

	if result.IsUp {
		t.Fatal("expected service down on connection refused, got up")
	}
}

func TestCheckTCP(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	c := NewChecker(2 * time.Second)

	result := c.check(context.Background(), Target{ServiceID: "svc-1", Type: "tcp", URL: l.Addr().String()})
	if !result.IsUp {
		t.Fatalf("expected TCP service up, got down: %s", result.ErrorMessage)
	}

	// Closed port
	closed, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := closed.Addr().String()
	closed.Close()

	result = c.check(context.Background(), Target{ServiceID: "svc-1", Type: "tcp", URL: addr, Timeout: time.Second})
	if result.IsUp {
		t.Fatal("expected TCP service down on closed port, got up")
	}
}

func TestCheckWithRetriesRecovers(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewChecker(5 * time.Second)
	result := c.checkWithRetries(context.Background(), Target{
		ServiceID: "svc-1",
		Type:      "http",
		URL:       srv.URL,
		Retries:   2,
	})

	if !result.IsUp {
		t.Fatalf("expected up after retry, got down: %s", result.ErrorMessage)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestCheckAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewChecker(5 * time.Second)
	targets := []Target{
		{ServiceID: "a", Type: "http", URL: srv.URL, Retries: 1},
		{ServiceID: "b", Type: "http", URL: srv.URL, Retries: 1},
		{ServiceID: "c", Type: "http", URL: srv.URL, Retries: 1},
	}

	results := c.CheckAll(context.Background(), targets)

	if len(results) != len(targets) {
		t.Fatalf("got %d results, want %d", len(results), len(targets))
	}
	for _, r := range results {
		if !r.IsUp {
			t.Errorf("service %s expected up: %s", r.ServiceID, r.ErrorMessage)
		}
	}
}
