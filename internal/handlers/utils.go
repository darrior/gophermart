package handlers

import "net/http"

type requestContextKey int

const userUUIDKey requestContextKey = 1

func setAuthCookie(w http.ResponseWriter, password, uuid string) error {
	panic("todo")
}

func validateLuhn(number string) error {
	panic("todo")
}
