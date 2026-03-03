package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

const userUUIDKey requestContextKey = 1

type requestContextKey int

type authClaims struct {
	jwt.RegisteredClaims
	UserUUID string
}

func (h *handlers) validateAuthCookie(ctx context.Context, tokenString string) (string, error) {
	claims := &authClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		c, ok := t.Claims.(*authClaims)
		if !ok {
			return nil, errors.New("cannot assert token claims")
		}

		passHash, err := h.s.GetPasswordHash(ctx, c.UserUUID)
		if err != nil {
			return nil, fmt.Errorf("cannot get password hash: %w", err)
		}

		return hex.DecodeString(passHash)
	})
	if err != nil {
		return "", fmt.Errorf("cannot parse JWT token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.UserUUID, nil
}

func setAuthCookie(w http.ResponseWriter, passHash, uuid string) error {
	tokenString, err := signClaims(passHash, uuid)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:  authCookieName,
		Value: tokenString,
	}

	http.SetCookie(w, cookie)

	return nil
}

func signClaims(passHash string, uuid string) (string, error) {
	hash, err := hex.DecodeString(passHash)
	if err != nil {
		return "", fmt.Errorf("cannot decode pass hash: %w", err)
	}

	claims := &authClaims{
		UserUUID: uuid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(hash)
	if err != nil {
		return "", fmt.Errorf("cannot sigh token: %w", err)
	}
	return tokenString, nil
}

func validateLuhn(number string) error {
	runes := []rune(number)
	slices.Reverse(runes)

	sum := 0
	for i, r := range runes {
		num, err := strconv.Atoi(string(r))
		if err != nil {
			return errors.New("non-digit character in number")
		}

		if i%2 == 0 {
			sum += num
			continue
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

func getUUIDFromContext(ctx context.Context) (string, error) {
	rawUUID := ctx.Value(userUUIDKey)
	if rawUUID == nil {
		return "", errors.New("cannot get UUID from context")
	}

	uuid, ok := rawUUID.(string)
	if !ok {
		return "", errors.New("cannot convert UUID to string")
	}

	return uuid, nil
}
