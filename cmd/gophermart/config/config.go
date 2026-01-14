// Package config provides abstractions over startup config.
package config

import "github.com/jackc/pgx/v5"

type Config struct {
	RunAddress           string          `env:"RUN_ADDRESS"`
	DatabaseURI          *pgx.ConnConfig `env:"DATABASE_URI"`
	AccrualSystemAddress string          `env:"ACCRUAL_SYSTEM"`
}

var DefaultConfig = Config{
	RunAddress:           ":8080",
	DatabaseURI:          nil,
	AccrualSystemAddress: "127.0.0.1:3000",
}

func ParseConfig() (Config, error) {
	return DefaultConfig, nil
}
