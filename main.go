package main

import (
	"context"
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
}

// NewService intializes an instance of the Service bound to a given address.
func NewService(ctx context.Context, addr string) (*Service, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", HelloWorldHandler)

	lc := net.ListenConfig{}
	l, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("couldn't listen on %q: %w", addr, err)
	}
	log.Printf("listening on %q\n", addr)

	return &Service{r: r, l: l}, nil
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

func main() {
	s, err := NewService(context.Background(), defaultAddress)
	if err != nil {
		log.Fatalf("couldn't start service: %b", err)
	}
	log.Fatal(s.Serve())
}
