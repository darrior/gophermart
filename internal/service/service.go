// Package service implements buisness-logic of app.
package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/darrior/gophermart/internal/gateways/accrual"
	"github.com/darrior/gophermart/internal/models"
	"github.com/darrior/gophermart/internal/repository"
	"github.com/darrior/gophermart/internal/utils/syncqueue"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	ErrLoginExists           = errors.New("login already exists")
	ErrInvalidCredentials    = errors.New("invaldi credentials")
	ErrOrderAlreadyExists    = errors.New("order already added")
	ErrOrderOwnedByOtherUser = errors.New("order was uploaded by other user")
	ErrInsufficientFunds     = errors.New("insufficient funds")
)

type Repository interface {
	AddUser(ctx context.Context, uuid, login, passHash string) (err error)
	AddOrder(ctx context.Context, uuid, number string, timestamp time.Time) (err error)
	AddWithdrawal(ctx context.Context, uuid, number string, balance, sum float64, time time.Time) (err error)
	UpdateOrderStatus(ctx context.Context, number string, status models.OrderStatus) (err error)
	UpdateOrder(ctx context.Context, order models.Order) (err error)
	GetOrder(ctx context.Context, number string) (order models.Order, err error)
	GetUser(ctx context.Context, uuid string) (user models.User, err error)
	GetUserByLogin(ctx context.Context, login string) (user models.User, err error)
	ListOrders(ctx context.Context, uuid string) (orders []models.Order, err error)
	ListWithdrawals(ctx context.Context, uuid string) (withdrawals []models.Withdrawal, err error)
}

type AccrualSystem interface {
	GetOrder(number string) (order models.AccrualOrderState, err error)
}

type service struct {
	repository    Repository
	accrualSystem AccrualSystem
	orderCfg      struct {
		orderWorkers int
		orderQueue   *syncqueue.SyncQueue[string]
		orderCond    *sync.Cond
	}
}

func NewService(ctx context.Context, r Repository, a AccrualSystem, workers int) *service {
	s := &service{
		repository:    r,
		accrualSystem: a,
		orderCfg: struct {
			orderWorkers int
			orderQueue   *syncqueue.SyncQueue[string]
			orderCond    *sync.Cond
		}{
			orderWorkers: workers,
			orderQueue:   syncqueue.NewSyncQueue[string](),
			orderCond:    &sync.Cond{},
		}}

	s.startWorkers(ctx)

	return s
}

func (s *service) RegisterUser(ctx context.Context, login, password string) (models.User, error) {
	if _, err := s.repository.GetUserByLogin(ctx, login); err == nil {
		return models.User{}, ErrLoginExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return models.User{}, fmt.Errorf("cannot get user: %w", err)
	}

	userUUID, err := uuid.NewRandom()
	if err != nil {
		return models.User{}, fmt.Errorf("cannot generate UUID: %w", err)
	}

	passHash := getPasswordHash(password)
	if err := s.repository.AddUser(ctx, userUUID.String(), login, passHash); err != nil {
		return models.User{}, fmt.Errorf("cannot add new user: %w", err)
	}

	return models.User{
		UUID:             userUUID.String(),
		Login:            login,
		PasswordHash:     passHash,
		CurrentBalance:   0,
		WithdrawnBalance: 0,
	}, nil
}

func (s *service) LoginUser(ctx context.Context, login, password string) (models.User, error) {
	user, err := s.repository.GetUserByLogin(ctx, login)
	if errors.Is(err, repository.ErrUserNotFound) {
		return models.User{}, ErrInvalidCredentials
	} else if err != nil {
		return models.User{}, fmt.Errorf("cannot get user: %w", err)
	}

	passHash := getPasswordHash(password)
	if user.PasswordHash != passHash {
		return models.User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (s *service) AddOrder(ctx context.Context, uuid, order string) error {
	var orderExists *repository.ErrorOrderExists
	if err := s.repository.AddOrder(ctx, uuid, order, time.Now()); errors.As(err, &orderExists) {
		if orderExists.UUID == uuid {
			return ErrOrderAlreadyExists
		} else {
			return ErrOrderOwnedByOtherUser
		}
	} else if err != nil {
		return fmt.Errorf("cannot add order to repository: %w", err)
	}

	s.pushOrder(order)

	return nil
}

func (s *service) Withdraw(ctx context.Context, uuid, order string, sum float64) error {
	user, err := s.repository.GetUser(ctx, uuid)
	if err != nil {
		return fmt.Errorf("cannot get user: %w", err)
	}

	if user.CurrentBalance < sum {
		return ErrInsufficientFunds
	}

	balance := user.CurrentBalance - sum
	if err := s.repository.AddWithdrawal(ctx, uuid, order, balance, sum, time.Now()); err != nil {
		return fmt.Errorf("cannot store withdraw: %w", err)
	}

	return nil
}

func (s *service) ListOrders(ctx context.Context, uuid string) ([]models.OrderResponse, error) {
	orders, err := s.repository.ListOrders(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("cannot get orders from repository: %w", err)
	}

	ret := []models.OrderResponse{}
	for _, order := range orders {
		ret = append(ret, models.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format(timeLayout),
		})
	}

	return ret, nil
}

func (s *service) ListWithdrawals(ctx context.Context, uuid string) ([]models.WithdrawResponse, error) {
	withdrawals, err := s.repository.ListWithdrawals(ctx, uuid)
	if err != nil {
		return []models.WithdrawResponse{}, fmt.Errorf("cannot get user from repository: %w", err)
	}

	ret := []models.WithdrawResponse{}
	for _, withdrawal := range withdrawals {
		ret = append(ret, models.WithdrawResponse{
			Order:       withdrawal.Order,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt.Format(timeLayout),
		})
	}

	return ret, nil
}

func (s *service) GetBalance(ctx context.Context, uuid string) (models.BalanceResponse, error) {
	user, err := s.repository.GetUser(ctx, uuid)
	if err != nil {
		return models.BalanceResponse{}, fmt.Errorf("cannot get user from repository: %w", err)
	}

	return models.BalanceResponse{
		Current:   user.CurrentBalance,
		Withdrawn: user.WithdrawnBalance,
	}, nil
}

func (s *service) GetPasswordHash(ctx context.Context, uuid string) (string, error) {
	user, err := s.repository.GetUser(ctx, uuid)
	if err != nil {
		return "", fmt.Errorf("cannot get user from repository: %w", err)
	}

	return user.PasswordHash, nil
}

func (s *service) pushOrder(number string) {
	s.orderCfg.orderQueue.Push(number)
	s.orderCfg.orderCond.Signal()
}

func (s *service) startWorkers(ctx context.Context) {
	go func() {
		t := time.NewTicker(60 * time.Second)

	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-t.C:
			}

			wg := sync.WaitGroup{}

			for range s.orderCfg.orderWorkers {
				wg.Add(1)

				go func() {
					s.workerOrder(ctx)
					wg.Done()
				}()
			}

			wg.Wait()
			t.Reset(60 * time.Second)
		}
	}()
}

func (s *service) updateOrder(ctx context.Context, order models.AccrualOrderState) error {
	switch order.Status {
	case models.AccrualOrderStatusRegistered:
		if err := s.repository.UpdateOrderStatus(ctx, order.Order, models.OrderStatusNew); err != nil {
			return fmt.Errorf("cannot update order status: %w", err)
		}
	case models.AccrualOrderStatusProcessing:
		if err := s.repository.UpdateOrderStatus(ctx, order.Order, models.OrderStatusProcessing); err != nil {
			return fmt.Errorf("cannot update order status: %w", err)
		}
	case models.AccrualOrderStatusInvalid:
		if err := s.repository.UpdateOrderStatus(ctx, order.Order, models.OrderStatusInvalid); err != nil {
			return fmt.Errorf("cannot update order status: %w", err)
		}
	case models.AccrualOrderStatusProcessed:
		if err := s.repository.UpdateOrder(ctx, models.Order{
			Accrual: order.Accrual,
			Number:  order.Order,
			Status:  models.OrderStatusProcessed,
		}); err != nil {
			return fmt.Errorf("cannot update order status: %w", err)
		}
	}

	return nil
}

func (s *service) workerOrder(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		number, ok := s.orderCfg.orderQueue.Pop()
		if !ok {
			s.orderCfg.orderCond.L.Lock()
			for s.orderCfg.orderQueue.IsEmpty() {
				s.orderCfg.orderCond.Wait()
			}
			s.orderCfg.orderCond.L.Unlock()
			continue
		}

		order, err := s.accrualSystem.GetOrder(number)
		if errors.Is(err, accrual.ErrTooManyRequests) {
			s.orderCfg.orderQueue.Push(number)
			break
		} else if errors.Is(err, accrual.ErrOrderIsNotExist) {
			s.orderCfg.orderQueue.Push(number)
			continue
		} else if err != nil {
			log.Error().Err(err).Msg("An error occured in accrual system")
			continue
		}

		if err := s.updateOrder(ctx, order); err != nil {
			log.Error().Err(err).Msg("Cannot update order")
		}
	}
}
