// Package handlers provides handlers for server's routes
package handlers

import (
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
	RegisterUser(login, password string) (uuid string, err error)
	LoginUser(login, password string) (uuid string, err error)
	AddOrder(uuid, order string) (err error)
	Withdraw(uuid, order string, sum int) (err error)
	ListOrders(uuid string) (orders []models.Order, err error)
	ListWithdrawals(uuid string) (withdrawals []models.Withdraw, err error)
	GetBalance(uuid string) (balance models.Balance, err error)
}

type handlers struct {
	s Service
}

func (h *handlers) postAPIUserRegister(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.Header.Get("content-type"), "application/json") {
		http.Error(w, "invalid content-type", http.StatusBadRequest)
		return
	}

	var data models.AuthenticationData
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

	uuid, err := h.s.RegisterUser(data.Login, data.Password)
	if errors.Is(err, service.ErrLoginExists) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := setAuthCookie(w, data.Password, uuid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) postAPIUserLogin(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.Header.Get("content-type"), "application/json") {
		http.Error(w, "invalid content-type", http.StatusBadRequest)
		return
	}

	var data models.AuthenticationData
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

	uuid, err := h.s.LoginUser(data.Login, data.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := setAuthCookie(w, data.Password, uuid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) postAPIUserOrders(w http.ResponseWriter, req *http.Request) {
	if !strings.HasSuffix(req.Header.Get("content-type"), "text/plain") {
		http.Error(w, "invalid content-type", http.StatusBadRequest)
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

	rawUUID := req.Context().Value(userUUIDKey)
	if rawUUID == nil {
		http.Error(w, "cannot find login", http.StatusInternalServerError)
		return
	}

	order := string(data)
	if err := validateLuhn(order); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	uuid := rawUUID.(string)
	if err := h.s.AddOrder(uuid, order); errors.Is(err, service.ErrOrderAlreadyExists) {
		w.WriteHeader(http.StatusOK)
		return
	} else if errors.Is(err, service.ErrOrderOwnedByOtherUser) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *handlers) getAPIUserOrders(w http.ResponseWriter, req *http.Request) {
	rawUUID := req.Context().Value(userUUIDKey)
	if rawUUID == nil {
		http.Error(w, "cannot find login", http.StatusInternalServerError)
		return
	}

	uuid := rawUUID.(string)
	orders, err := h.s.ListOrders(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataLength := strconv.Itoa(len(data))

	w.Header().Set("content-type", "application/json")
	w.Header().Set("content-length", dataLength)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *handlers) getAPIUserBalance(w http.ResponseWriter, req *http.Request) {
	rawUUID := req.Context().Value(userUUIDKey)
	if rawUUID == nil {
		http.Error(w, "cannot find login", http.StatusInternalServerError)
		return
	}

	uuid := rawUUID.(string)
	balance, err := h.s.GetBalance(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataLength := strconv.Itoa(len(data))

	w.Header().Set("content-type", "application/json")
	w.Header().Set("content-length", dataLength)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *handlers) postAPIUserBalanceWithdrew(w http.ResponseWriter, req *http.Request) {
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

	rawUUID := req.Context().Value(userUUIDKey)
	if rawUUID == nil {
		http.Error(w, "cannot find login", http.StatusInternalServerError)
		return
	}

	uuid := rawUUID.(string)
	if err := h.s.Withdraw(uuid, data.Order, data.Sum); errors.Is(err, service.ErrInsufficientFunds) {
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) getAPIUserWithdrawals(w http.ResponseWriter, req *http.Request) {
	rawUUID := req.Context().Value(userUUIDKey)
	if rawUUID == nil {
		http.Error(w, "cannot find login", http.StatusInternalServerError)
		return
	}

	uuid := rawUUID.(string)
	withdrawals, err := h.s.ListWithdrawals(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(withdrawals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataLength := strconv.Itoa(len(data))

	w.Header().Set("content-type", "application/json")
	w.Header().Set("content-length", dataLength)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
