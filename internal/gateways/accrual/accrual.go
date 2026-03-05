// Package accrual provides methods to interact with accural system.
package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/darrior/gophermart/internal/models"
	"github.com/darrior/gophermart/internal/utils/client"
	"github.com/rs/zerolog/log"
)

var (
	ErrOrderIsNotExist = errors.New("order is not registered yet")
	ErrServerError     = errors.New("internal server error")
)

type ErrorTooManyRequests struct {
	RetryAfter time.Duration
}

func newErrorTooManyRequests(duration time.Duration) ErrorTooManyRequests {
	return ErrorTooManyRequests{
		RetryAfter: duration,
	}
}

func (e ErrorTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests; retry after %d", e.RetryAfter)
}

type getter interface {
	Get(url string) (*http.Response, error)
}

type accrual struct {
	client  getter
	baseURL string
}

func NewAccrual(address string) *accrual {
	return &accrual{
		client:  client.NewRetryableClient(),
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
		retry, err := strconv.Atoi(resp.Header.Get("retry-after"))
		if err != nil {
			return models.AccrualOrderState{}, newErrorTooManyRequests(time.Minute)
		}
		return models.AccrualOrderState{}, newErrorTooManyRequests(time.Duration(retry) * time.Second)
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
