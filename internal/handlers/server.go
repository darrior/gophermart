package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type server struct {
	h      *handlers
	mux    *chi.Mux
	server *http.Server
}

func NewServer(address string, service Service) *server {
	s := &server{
		mux: chi.NewMux(),
	}

	s.h = &handlers{service}
	s.setRoutes()
	server := &http.Server{
		Addr:    address,
		Handler: s.mux,
	}
	s.server = server

	return s
}

func (s *server) Start() error {
	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server unexpected exited: %w", err)
	}

	return nil
}

func (s *server) Stop() error {
	return s.server.Close()
}
