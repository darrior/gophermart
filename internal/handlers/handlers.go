// Package handlers provides handlers for server's routes
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/darrior/gophermart/internal/models"
	"github.com/darrior/gophermart/internal/service"
	"github.com/rs/zerolog/log"
)

type Service interface {
	RegisterUser(ctx context.Context, login, password string) (user models.User, err error)
	LoginUser(ctx context.Context, login, password string) (user models.User, err error)
	AddOrder(ctx context.Context, uuid, order string) (err error)
	Withdraw(ctx context.Context, uuid, order string, sum float64) (err error)
	ListOrders(ctx context.Context, uuid string) (orders []models.OrderResponse, err error)
	ListWithdrawals(ctx context.Context, uuid string) (withdrawals []models.WithdrawResponse, err error)
	GetBalance(ctx context.Context, uuid string) (balance models.BalanceResponse, err error)
	GetPasswordHash(ctx context.Context, uuid string) (passHash string, err error)
}

type handlers struct {
	s Service
}

func (h *handlers) postAPIUserRegister(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.Header.Get("content-type"), "application/json") {
		http.Error(w, "invalid content-type", http.StatusBadRequest)
		return
	}

	var data models.AuthenticationDataRequest
	d := json.NewDecoder(req.Body)
	if err := d.Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Can not close request body")
		}
	}()

	ctx := req.Context()
	user, err := h.s.RegisterUser(ctx, data.Login, data.Password)
	if errors.Is(err, service.ErrLoginExists) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := setAuthCookie(w, user.PasswordHash, user.UUID); err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) postAPIUserLogin(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.Header.Get("content-type"), "application/json") {
		http.Error(w, "invalid content-type", http.StatusBadRequest)
		return
	}

	var data models.AuthenticationDataRequest
	d := json.NewDecoder(req.Body)
	if err := d.Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Can not close request body")
		}
	}()

	ctx := req.Context()
	user, err := h.s.LoginUser(ctx, data.Login, data.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := setAuthCookie(w, user.PasswordHash, user.UUID); err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) postAPIUserOrders(w http.ResponseWriter, req *http.Request) {
	if !strings.HasSuffix(req.Header.Get("content-type"), "text/plain") {
		http.Error(w, "invalid content-type", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	uuid, err := getUUIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Can not close request body")
		}
	}()

	order := strings.TrimSpace(string(data))
	log.Info().Str("order", order).Msg("Order received")
	if err := validateLuhn(order); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.s.AddOrder(ctx, uuid, order); errors.Is(err, service.ErrOrderAlreadyExists) {
		w.WriteHeader(http.StatusOK)
		return
	} else if errors.Is(err, service.ErrOrderOwnedByOtherUser) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *handlers) getAPIUserOrders(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	uuid, err := getUUIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	orders, err := h.s.ListOrders(ctx, uuid)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(orders)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	dataLength := strconv.Itoa(len(data))

	w.Header().Set("content-type", "application/json")
	w.Header().Set("content-length", dataLength)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *handlers) getAPIUserBalance(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	uuid, err := getUUIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	balance, err := h.s.GetBalance(ctx, uuid)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(balance)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	dataLength := strconv.Itoa(len(data))

	w.Header().Set("content-type", "application/json")
	w.Header().Set("content-length", dataLength)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *handlers) postAPIUserBalanceWithdraw(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.Header.Get("content-type"), "application/json") {
		http.Error(w, "invalid content-type", http.StatusBadRequest)
		return
	}

	var data models.WithdrawRequest
	d := json.NewDecoder(req.Body)
	if err := d.Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Can not close request body")
		}
	}()

	if err := validateLuhn(data.Order); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ctx := req.Context()
	uuid, err := getUUIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := h.s.Withdraw(ctx, uuid, data.Order, data.Sum); errors.Is(err, service.ErrInsufficientFunds) {
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	} else if err != nil {
		log.Error().Err(err).Msg("Cannot make withdraw")
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) getAPIUserWithdrawals(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	uuid, err := getUUIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	withdrawals, err := h.s.ListWithdrawals(ctx, uuid)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(withdrawals)
	if err != nil {
		log.Error().Err(err).Msg("Internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	dataLength := strconv.Itoa(len(data))

	w.Header().Set("content-type", "application/json")
	w.Header().Set("content-length", dataLength)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
