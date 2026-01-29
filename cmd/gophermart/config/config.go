// Package config provides abstractions over startup config.
package config

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5"
)

type Config struct {
	RunAddress           string
	DatabaseConnConf     *pgx.ConnConfig
	AccrualSystemAddress string
}

type rawConfig struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM"`
}

var DefaultConfig = rawConfig{
	RunAddress:           ":8080",
	DatabaseURI:          "",
	AccrualSystemAddress: "127.0.0.1:3000",
}

func ParseConfig() (Config, error) {
	r := DefaultConfig

	flags := flag.NewFlagSet("default flags", flag.ContinueOnError)
	flags.StringVar(&r.RunAddress, "a", r.RunAddress, "service address in format host:port")
	flags.StringVar(&r.DatabaseURI, "d", r.DatabaseURI, "DB address")
	flags.StringVar(&r.AccrualSystemAddress, "r", r.AccrualSystemAddress, "accrual service address")

	if err := flags.Parse(os.Args[1:]); err != nil {
		return Config{}, fmt.Errorf("cannot parse CLI arguments: %w", err)
	}

	if err := env.Parse(&r); err != nil {
		return Config{}, fmt.Errorf("cannot parse env: %w", err)
	}

	return validateConfig(r)
}

func validateConfig(r rawConfig) (Config, error) {
	runAddr, err := validateRunAddr(r.RunAddress)
	if err != nil {
		return Config{}, fmt.Errorf("run address is invalid: %w", err)
	}

	dbConnConf, err := validateDBURI(r.DatabaseURI)
	if err != nil {
		return Config{}, fmt.Errorf("DB URI is invalid: %w", err)
	}

	accrualAddr, err := ValidateAccrualAddr(r.AccrualSystemAddress)
	if err != nil {
		return Config{}, fmt.Errorf("accrual address is invalid: %w", err)
	}

	return Config{
		RunAddress:           runAddr,
		DatabaseConnConf:     dbConnConf,
		AccrualSystemAddress: accrualAddr,
	}, nil
}

func validateRunAddr(address string) (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return "", fmt.Errorf("cannot resolve address: %w", err)
	}

	return addr.String(), nil
}

func validateDBURI(dbURI string) (*pgx.ConnConfig, error) {
	connConf, err := pgx.ParseConfig(dbURI)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DB URI: %w", err)
	}

	return connConf, nil
}

func ValidateAccrualAddr(accrualAddr string) (string, error) {
	url, err := url.Parse(accrualAddr)
	if err != nil {
		return "", fmt.Errorf("cannot parse url: %w", err)
	}

	return url.String(), nil
}
