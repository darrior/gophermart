package service

import (
	"crypto/sha256"
)

const timeLayout = "2006-01-02T15:04:05-07:00"

func getPasswordHash(password string) string {
	return string(sha256.New().Sum([]byte(password)))
}
