// Package service implements buisness-logic of app.
package service

import "errors"

var (
	ErrLoginExists           = errors.New("login already exists")
	ErrInvalidCredentials    = errors.New("invaldi credentials")
	ErrOrderAlreadyExists    = errors.New("order already added")
	ErrOrderOwnedByOtherUser = errors.New("order was uploaded by other user")
	ErrInsufficientFunds     = errors.New("insufficient funds")
)
