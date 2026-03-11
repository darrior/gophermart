package service

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// const timeLayout = "2006-01-02T15:04:05-07:00"
const timeLayout = time.RFC3339

func getPasswordHash(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}
