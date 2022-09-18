package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultURL string = fmt.Sprintf("http://%s/", defaultAddress)

func TestServiceServe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the service in the background
	s, err := NewService(ctx, defaultAddress)
	if err != nil {
		t.Fatalf("failed to start service: %v", err)
	}
	errCh := make(chan error)
	go func() {
		errCh <- s.Serve()
		close(errCh)
	}()

	// Check to make sure serve didn't instantly fail
	select {
	case err := <-errCh:
		t.Fatalf("error while serving: %v", err)
	default:
	}

	// Send a GET to the base url
	resp, err := http.Get(defaultURL)
	if err != nil {
		t.Fatalf("GET request could not be completed: %v", err)
	}
	defer resp.Body.Close()

	// Verify response
	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if got, want := string(body), "Hello World!"; got != want {
		t.Errorf("returned unexpected body: got %v want %v", got, want)
	}
}

// TestHelloWorldHandler checks the default handler for a simple "Hello World!"
//
//	response.
func TestHelloWorldHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("couldn't build new request: %s", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HelloWorldHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if got, want := rr.Body.String(), "Hello World!"; got != want {
		t.Errorf("returned unexpected body: got %v want %v", got, want)
	}
}
