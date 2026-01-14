// Package accrual provides methods to interact with accural system.
package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/darrior/gophermart/internal/models"
	"github.com/rs/zerolog/log"
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrOrderIsNotExist = errors.New("order is not registered yet")
	ErrServerError     = errors.New("internal server error")
)

type accrual struct {
	client  *http.Client
	baseURL string
}

func NewAccrual(address string) *accrual {
	return &accrual{
		client:  &http.Client{},
		baseURL: address,
	}
}

func (a *accrual) GetOrder(number string) (models.AccrualOrderState, error) {
	url, err := url.JoinPath(a.baseURL, "api/orders", number)
	if err != nil {
		return models.AccrualOrderState{}, fmt.Errorf("cannot create request URL: %w", err)
	}

	resp, err := a.client.Get(url)
	if err != nil {
		return models.AccrualOrderState{}, fmt.Errorf("cannot perform request: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return models.AccrualOrderState{}, ErrOrderIsNotExist
	case http.StatusTooManyRequests:
		return models.AccrualOrderState{}, ErrTooManyRequests
	case http.StatusInternalServerError:
		return models.AccrualOrderState{}, ErrServerError
	}

	d := json.NewDecoder(resp.Body)
	var order models.AccrualOrderState
	if err := d.Decode(&order); err != nil {
		return models.AccrualOrderState{}, fmt.Errorf("cannot unmarshal response: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Cannot close response body")
		}
	}()

	return order, nil
}
