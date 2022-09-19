package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
// response.
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

func TestFlowsPOST(t *testing.T) {
	tcs := []struct {
		desc string
		json string
		hour int
		want []Flow
	}{
		{
			"first example from spec",
			`[
			{
			   "src_app":"foo", "dest_app":"bar", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":300, "hour":1
			},
			{
			   "src_app":"foo", "dest_app":"bar", "vpc_id":"vpc-0",
			   "bytes_tx":200, "bytes_rx":600, "hour":1
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":500, "hour":1
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":500, "hour":2
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-1",
			   "bytes_tx":100, "bytes_rx":500, "hour":2
			}
		 ]`,
			1,
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 300, 900),
				NewFlow("vpc-0", "baz", "qux", 1, 100, 500),
			},
		},
		{
			"second example from spec",
			`[
			{
			   "src_app":"foo", "dest_app":"bar", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":300, "hour":1
			},
			{
			   "src_app":"foo", "dest_app":"bar", "vpc_id":"vpc-0",
			   "bytes_tx":200, "bytes_rx":600, "hour":1
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":500, "hour":1
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":500, "hour":2
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-1",
			   "bytes_tx":100, "bytes_rx":500, "hour":2
			}
		 ]`,
			2,
			[]Flow{
				NewFlow("vpc-0", "baz", "qux", 2, 100, 500),
				NewFlow("vpc-1", "baz", "qux", 2, 100, 500),
			},
		},
		{
			"third example from spec",
			`[
			{
			   "src_app":"foo", "dest_app":"bar", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":300, "hour":1
			},
			{
			   "src_app":"foo", "dest_app":"bar", "vpc_id":"vpc-0",
			   "bytes_tx":200, "bytes_rx":600, "hour":1
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":500, "hour":1
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-0",
			   "bytes_tx":100, "bytes_rx":500, "hour":2
			},
			{
			   "src_app":"baz", "dest_app":"qux", "vpc_id":"vpc-1",
			   "bytes_tx":100, "bytes_rx":500, "hour":2
			}
		 ]`,
			3,
			[]Flow{},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			s := &Service{
				c: NewMemoryController(),
			}

			req, err := http.NewRequestWithContext(ctx, "POST", "/flows", strings.NewReader(tc.json))
			if err != nil {
				t.Fatalf("couldn't build new request: %s", err)
			}
			req.Header.Add("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handlerWithSvc(s, FlowsPOST))
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			got, err := s.c.FlowReadHour(ctx, tc.hour)
			if err != nil {
				t.Fatalf("error during FlowReadHour: %v", err)
			}

			if !equalsUnordered(tc.want, got) {
				t.Fatalf("ReadAll returned unexpected result: got %v, want %v", got, tc.want)
			}
		})
	}
}
