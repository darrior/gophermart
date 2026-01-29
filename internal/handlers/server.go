package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
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

func (s *server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		if err := s.server.Close(); err != nil {
			log.Error().Err(err).Msg("Cannot stop server properly")
		}
	}()

	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server unexpected exited: %w", err)
	}

	return nil
}
