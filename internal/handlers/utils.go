package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

const userUUIDKey requestContextKey = 1

type requestContextKey int

type authClaims struct {
	jwt.Claims
	userUUID string
}

func (h *handlers) validateAuthCookie(ctx context.Context, tokenString string) (string, error) {
	var claims authClaims

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		c, ok := t.Claims.(authClaims)
		if !ok {
			return nil, errors.New("cannot assert token claims")
		}

		return h.s.GetPasswordHash(ctx, c.userUUID)
	})
	if err != nil {
		return "", fmt.Errorf("cannot parse JWT token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.userUUID, nil
}

func setAuthCookie(w http.ResponseWriter, passHash, uuid string) error {
	claims := &authClaims{
		userUUID: uuid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(passHash)
	if err != nil {
		return fmt.Errorf("cannot sigh token: %w", err)
	}

	cookie := &http.Cookie{
		Name:  authCookieName,
		Value: tokenString,
	}

	http.SetCookie(w, cookie)

	return nil
}

func validateLuhn(number string) error {
	runes := []rune(number)
	slices.Reverse(runes)

	sum := 0
	for i, r := range runes {
		if i%2 == 0 {
			sum += i
			continue
		}

		num, err := strconv.Atoi(string(r))
		if err != nil {
			return errors.New("non-digit character in number")
		}

		num = num * 2
		if num > 9 {
			num -= 9
		}

		sum += num
	}

	if sum%10 != 0 {
		return errors.New("number is invalid")
	}

	return nil
}
