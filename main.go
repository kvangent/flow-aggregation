package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

const defaultAddress string = "localhost:8080"

// Service represents an instance of the service. Use NewService to intialize.
type Service struct {
	r *mux.Router
	l net.Listener
	c Controller
}

// NewService intializes an instance of the Service bound to a given address.
func NewService(ctx context.Context, addr string) (*Service, error) {
	s := &Service{}
	s.r = mux.NewRouter()
	s.r.HandleFunc("/", HelloWorldHandler)
	s.r.Path("/flows").Methods("POST").HandlerFunc(handlerWithSvc(s, FlowsPOST))

	lc := net.ListenConfig{}
	var err error
	s.l, err = lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("couldn't listen on %q: %w", addr, err)
	}
	log.Printf("listening on %q\n", addr)

	return s, nil
}

// Serve causes an Service to serve HTTP requests until an error is reached.
func (a *Service) Serve() error {
	return http.Serve(a.l, a.r)
}

// HelloWorldHandler responds "Hello World!"
func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/*")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, "Hello World!")
}

// FlowsPOST takes an array of Flows in JSON format, and aggregates their use
// into the background controller.
func FlowsPOST(s *Service, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// verify the content type should be JSON
	if c := r.Header.Get("Content-Type"); c != "application/json" {
		log.Printf("error: 'Content-Type was '%s'\n", c)
		http.Error(w, "'Content-Type' must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// read the entire body
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		// Note: for a more helpful but less secure interface, we can return
		// some of these errors to the end user instead of only logging.
		log.Printf("error: unable to read request body: %v\n", err)
		http.Error(w, "error reading request body", http.StatusBadRequest)
		return
	}

	// deserialze the body
	var f []Flow
	if err := json.Unmarshal(b, &f); err != nil {
		log.Printf("error: unable to deseralize json: %v", err)
		http.Error(w, "body was incorrect or invalid JSON", http.StatusBadRequest)
		return
	}

	// commit flows using the controller
	if err := s.c.FlowAggregate(ctx, f); err != nil {
		log.Printf("error: unable to update flows in datasource: %v", err)
		http.Error(w, "server errro while aggregating flows", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handlerWithSvc is a helper function to convert a handler that takes a service
// into the more traditional func signiture.
func handlerWithSvc(s *Service, f func(s *Service, w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(s, w, r)
	}
}

func main() {
	s, err := NewService(context.Background(), defaultAddress)
	if err != nil {
		log.Fatalf("couldn't start service: %b", err)
	}
	log.Fatal(s.Serve())
}
