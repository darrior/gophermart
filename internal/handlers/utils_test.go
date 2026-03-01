package handlers

import (
	"context"
	"encoding/hex"
	"net/http/httptest"
	"testing"

	"github.com/darrior/gophermart/internal/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_handlers_validateAuthCookie(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := mocks.NewMockService(ctrl)

	s.EXPECT().
		GetPasswordHash(gomock.Any(), "eaa33405-af8b-4eab-9c52-0d1c6ee7fd17").
		Return("a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3", nil)

	key, _ := hex.DecodeString("a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3")
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&authClaims{
			RegisteredClaims: jwt.RegisteredClaims{},
			UserUUID:         "eaa33405-af8b-4eab-9c52-0d1c6ee7fd17"},
	)
	tokenString, _ := token.SignedString(key)

	type fields struct {
		s Service
	}
	type args struct {
		ctx         context.Context
		tokenString string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "",
			fields: fields{
				s: s,
			},
			args: args{
				ctx:         context.TODO(),
				tokenString: tokenString,
			},
			want:      "eaa33405-af8b-4eab-9c52-0d1c6ee7fd17",
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handlers{
				s: tt.fields.s,
			}
			got, err := h.validateAuthCookie(tt.args.ctx, tt.args.tokenString)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_setAuthCookie(t *testing.T) {
	type args struct {
		w        *httptest.ResponseRecorder
		passHash string
		uuid     string
	}
	tests := []struct {
		name      string
		args      args
		want      int
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Valid test",
			args: args{
				w:        &httptest.ResponseRecorder{},
				passHash: "181210f8f9c779c26da1d9b2075bde0127302ee0e3fca38c9a83f5b1dd8e5d3b",
				uuid:     "cbae8147-ceef-452f-9ff5-e477d9c85e2f",
			},
			want:      1,
			assertion: assert.NoError,
		},
		{
			name: "Invalid hash",
			args: args{
				w:        &httptest.ResponseRecorder{},
				passHash: "1asdfxz",
				uuid:     "cbae8147-ceef-452f-9ff5-e477d9c85e2f",
			},
			want:      0,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, setAuthCookie(tt.args.w, tt.args.passHash, tt.args.uuid))

			response := tt.args.w.Result()
			defer func() { _ = response.Body.Close() }()

			assert.Equal(t, tt.want, len(response.Cookies()))
		})
	}
}

func Test_validateLuhn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		asserion assert.ErrorAssertionFunc
	}{
		{
			name:     "Valid Luhn number",
			input:    "79927398713",
			asserion: assert.NoError,
		},
		{
			name:     "Invalid Luhn number",
			input:    "78927398713",
			asserion: assert.Error,
		},
		{
			name:     "Non-digit character",
			input:    "79927E98713",
			asserion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLuhn(tt.input)
			tt.asserion(t, err)
		})
	}
}

func Test_signClaims(t *testing.T) {
	type args struct {
		passHash string
		uuid     string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "",
			args: args{
				passHash: "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
				uuid:     "b55d1e7c-daca-4f75-abda-018cf107add8",
			},
			want:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VyVVVJRCI6ImI1NWQxZTdjLWRhY2EtNGY3NS1hYmRhLTAxOGNmMTA3YWRkOCJ9.ZWqYxh8Wi8CRJH7oGqWPGf0RC7_yItTURBYQcu7MC2w",
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := signClaims(tt.args.passHash, tt.args.uuid)

			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
