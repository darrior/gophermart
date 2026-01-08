package handlers

import "github.com/go-chi/chi/v5"

type server struct {
	h       *handlers
	mux     chi.Router
	address string
}
