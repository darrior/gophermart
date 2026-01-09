// Package repository provides abstractions over DB storage.
package repository

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrOrderNotFound = errors.New("order not found")
)

type repository struct {
}
