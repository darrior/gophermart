// Package config provides abstractions over startup config.
package config

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func Test_validateConfig(t *testing.T) {
	type args struct {
		r rawConfig
	}
	tests := []struct {
		name      string
		args      args
		want      Config
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Valid test",
			args: args{
				r: rawConfig{
					RunAddress:           ":8080",
					DatabaseURI:          "postgres://postgres:123@example.com:5432/postgres?sslmode=disable",
					AccrualSystemAddress: "http://127.0.0.1:3000",
				},
			},
			want: Config{
				RunAddress: ":8080",
				DatabaseConnConf: func() *pgx.ConnConfig {
					c, _ := pgx.ParseConfig("postgres://postgres:123@example.com:5432/postgres?sslmode=disable")
					return c
				}(),
				AccrualSystemAddress: "http://127.0.0.1:3000",
			},
			assertion: assert.NoError,
		},
		{
			name: "Invalid run address",
			args: args{
				r: rawConfig{
					RunAddress:           "15",
					DatabaseURI:          "postgres://postgres:123@example.com:5432/postgres?sslmode=disable",
					AccrualSystemAddress: "http://127.0.0.1:3000",
				},
			},
			want: Config{
				RunAddress:           "",
				DatabaseConnConf:     nil,
				AccrualSystemAddress: "",
			},
			assertion: assert.Error,
		},
		{
			name: "Invalid DSN",
			args: args{
				r: rawConfig{
					RunAddress:           ":8080",
					DatabaseURI:          "example.com",
					AccrualSystemAddress: "",
				},
			},
			want: Config{
				RunAddress:           "",
				DatabaseConnConf:     nil,
				AccrualSystemAddress: "",
			},
			assertion: assert.Error,
		},
		{
			name: "Invalid accrual system address",
			args: args{
				r: rawConfig{
					RunAddress:           ":8080",
					DatabaseURI:          "postgres://postgres:123@example.com:5432/postgres?sslmode=disable",
					AccrualSystemAddress: "127.0.0.1:3000",
				},
			},
			want: Config{
				RunAddress:           "",
				DatabaseConnConf:     nil,
				AccrualSystemAddress: "",
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateConfig(tt.args.r)
			tt.assertion(t, err)

			assert.Equal(t, tt.want.AccrualSystemAddress, got.AccrualSystemAddress)
			assert.Equal(t, tt.want.RunAddress, got.RunAddress)
		})
	}
}

func Test_validateRunAddr(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Localhost",
			args: args{
				address: "127.0.0.1:8080",
			},
			want:      "127.0.0.1:8080",
			assertion: assert.NoError,
		},
		{
			name: "Address without port",
			args: args{
				address: "10.10.10.1",
			},
			want:      "",
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateRunAddr(tt.args.address)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_validateDBURI(t *testing.T) {
	type args struct {
		dbURI string
	}
	tests := []struct {
		name string
		args args

		wantAssertion assert.ValueAssertionFunc
		errAssertion  assert.ErrorAssertionFunc
	}{
		{
			name: "Valid test",
			args: args{
				dbURI: "postgres://postgres:123@example.com:5432/postgres?sslmode=disable",
			},
			wantAssertion: assert.NotNil,
			errAssertion:  assert.NoError,
		},
		{
			name: "Invalid test",
			args: args{
				dbURI: "example.com",
			},
			wantAssertion: assert.Nil,
			errAssertion:  assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateDBURI(tt.args.dbURI)

			tt.errAssertion(t, err)
			tt.wantAssertion(t, got)
		})
	}
}

func Test_validateAccrualAddr(t *testing.T) {
	type args struct {
		accrualAddr string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Valid test with domain",
			args: args{
				accrualAddr: "http://example.com",
			},
			want:      "http://example.com",
			assertion: assert.NoError,
		},
		{
			name: "Valid test with IP",
			args: args{
				accrualAddr: "https://127.0.0.1:9010",
			},
			want:      "https://127.0.0.1:9010",
			assertion: assert.NoError,
		},
		{
			name: "Without scheme",
			args: args{
				accrualAddr: "127.0.0.1:123",
			},
			want:      "",
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateAccrualAddr(tt.args.accrualAddr)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
