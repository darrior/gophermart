// Package service implements buisness-logic of app.
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/darrior/gophermart/internal/models"
	"github.com/darrior/gophermart/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrLoginExists           = errors.New("login already exists")
	ErrInvalidCredentials    = errors.New("invaldi credentials")
	ErrOrderAlreadyExists    = errors.New("order already added")
	ErrOrderOwnedByOtherUser = errors.New("order was uploaded by other user")
	ErrInsufficientFunds     = errors.New("insufficient funds")
)

type Repository interface {
	AddUser(uuid, login, passHash string) (err error)
	AddOrder(uuid, number string) (err error)
	AddWithdrawal(uuid, number string, balance, sum float64, time time.Time) (err error)
	GetOrder(number string) (order models.Order, err error)
	GetUser(uuid string) (user models.User, err error)
	GetUserByLogin(login string) (user models.User, err error)
	ListOrders(uuid string) (orders []models.Order, err error)
	ListWithdrawals(uuid string) (withdrawals []models.Withdrawal, err error)
}

type service struct {
	repository Repository
}

func NewService(r Repository) *service {
	return &service{repository: r}
}

func (s *service) RegisterUser(login, password string) (string, error) {
	if _, err := s.repository.GetUserByLogin(login); err == nil {
		return "", ErrLoginExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return "", fmt.Errorf("cannot get user: %w", err)
	}

	userUUID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate UUID: %w", err)
	}

	passHash := getPasswordHash(password)
	if err := s.repository.AddUser(userUUID.String(), login, passHash); err != nil {
		return "", fmt.Errorf("cannot add new user: %w", err)
	}

	return userUUID.String(), nil
}

func (s *service) LoginUser(login, password string) (string, error) {
	user, err := s.repository.GetUserByLogin(login)
	if errors.Is(err, repository.ErrUserNotFound) {
		return "", ErrInvalidCredentials
	}

	passHash := getPasswordHash(password)
	if user.PasswordHash != passHash {
		return "", ErrInvalidCredentials
	}

	return user.UUID, nil
}

func (s *service) AddOrder(uuid, order string) error {
	if o, err := s.repository.GetOrder(order); err == nil {
		if o.UserUUID == uuid {
			return ErrOrderAlreadyExists
		}
		return ErrOrderOwnedByOtherUser
	} else if !errors.Is(err, repository.ErrOrderNotFound) {
		return fmt.Errorf("cannot get order from repository: %w", err)
	}

	if err := s.repository.AddOrder(uuid, order); err != nil {
		return fmt.Errorf("cannot add order to repository: %w", err)
	}

	return nil
}

func (s *service) Withdraw(uuid, order string, sum float64) error {
	user, err := s.repository.GetUser(uuid)
	if err != nil {
		return fmt.Errorf("cannot get user: %w", err)
	}

	if user.CurrentBalance < sum {
		return ErrInsufficientFunds
	}

	balance := user.CurrentBalance - sum
	if err := s.repository.AddWithdrawal(uuid, order, balance, sum, time.Now()); err != nil {
		return fmt.Errorf("cannot store withdraw: %w", err)
	}

	return nil
}

func (s *service) ListOrders(uuid string) ([]models.OrderResponse, error) {
	orders, err := s.repository.ListOrders(uuid)
	if err != nil {
		return nil, fmt.Errorf("cannot get orders from repository: %w", err)
	}

	var ret []models.OrderResponse
	for _, order := range orders {
		ret = append(ret, models.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accural,
			UploadedAt: order.UploadedAt.Format(timeLayout),
		})
	}

	return ret, nil
}

func (s *service) ListWithdrawals(uuid string) ([]models.WithdrawResponse, error) {
	withdrawals, err := s.repository.ListWithdrawals(uuid)
	if err != nil {
		return nil, fmt.Errorf("cannot get user from repository: %w", err)
	}

	var ret []models.WithdrawResponse
	for _, withdrawal := range withdrawals {
		ret = append(ret, models.WithdrawResponse{
			Order:       withdrawal.Order,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt.Format(timeLayout),
		})
	}

	return ret, nil
}

func (s *service) GetBalance(uuid string) (models.BalanceResponse, error) {
	user, err := s.repository.GetUser(uuid)
	if err != nil {
		return models.BalanceResponse{}, fmt.Errorf("cannot get user from repository: %w", err)
	}

	return models.BalanceResponse{
		Current:   user.CurrentBalance,
		Withdrawn: user.WithdrawnBalance,
	}, nil
}
