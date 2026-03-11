package handlers

import (
	"context"
	"net/http"
)

const authCookieName = "auth"

func (h *handlers) authMiddlware(handler http.Handler) http.Handler {
	authHandler := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(authCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userUUID, err := h.validateAuthCookie(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		newReq := r.WithContext(context.WithValue(r.Context(), userUUIDKey, userUUID))

		handler.ServeHTTP(w, newReq)

	}

	return http.HandlerFunc(authHandler)
}
