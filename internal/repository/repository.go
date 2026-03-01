// Package repository provides abstractions over DB storage.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darrior/gophermart/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrOrderNotFound = errors.New("order not found")
	ErrOrderExists   = &ErrorOrderExists{}
)

type ErrorOrderExists struct {
	UUID string
}

func (e *ErrorOrderExists) Error() string {
	return fmt.Sprintf("order already uploaded by user %s", e.UUID)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) AddUser(ctx context.Context, uuid, login, passHash string) error {
	log.Info().Str("login", login).Msg("add user with login")

	if _, err := r.db.ExecContext(ctx, "INSERT INTO users (id, login, password_hash, current_balance, withdrawan_balance) VALUES ($1, $2, $3, $4, $5)", uuid, login, passHash, 0.0, 0.0); err != nil {
		return fmt.Errorf("cannot insert user: %w", err)
	}

	return nil
}

func (r *repository) AddOrder(ctx context.Context, uuid, order string, timestamp time.Time) error {
	row := r.db.QueryRowContext(ctx, "INSERT INTO orders (number, user_uuid, uploaded_at) VALUES ($1, $2, $3) ON CONFLICT DO UPDATE SET number = $1 RETURNING user_uuid, uploaded_at", order, uuid, timestamp)

	var (
		userUUID   string
		uploadedAt time.Time
	)

	if err := row.Scan(&userUUID, &uploadedAt); err != nil {
		return fmt.Errorf("cannot parse row: %w", err)
	}

	if uploadedAt.Before(timestamp) || userUUID != uuid {
		return &ErrorOrderExists{UUID: userUUID}
	}

	return nil
}

func (r *repository) AddWithdrawal(ctx context.Context, uuid, number string, balance, sum float64, time time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "UPDATE users SET current_balance = $1, withdrawal_balance = withdrawal_balance + $3 WHERE id = $2", balance, uuid, sum); err != nil {
		return fmt.Errorf("cannot update user: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO withdrawals (user_uuid, order, sum, processed_at) VALUES ($1, $2, $3, $4)", uuid, number, sum, time); err != nil {
		return fmt.Errorf("cannot add withdrawal: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("cannot complete transaction: %w", err)
	}

	return nil
}

func (r *repository) UpdateOrderStatus(ctx context.Context, number string, status models.OrderStatus) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE orders SET status = $2 WHERE id = $1", number, status); err != nil {
		return fmt.Errorf("cannot perform update: %w", err)
	}

	return nil
}

func (r *repository) UpdateOrder(ctx context.Context, order models.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction: %w", err)
	}

	row := tx.QueryRowContext(ctx, "UPDATE orders SET status = $2, accrual = $3 WHERE number = $1 RETURNING user_uuid", order.Number, order.Status, order.Accrual)

	if err := row.Err(); err != nil {
		return fmt.Errorf("cannot perform orders update: %w", err)
	}

	var userUUID string
	if err := row.Scan(&userUUID); err != nil {
		return fmt.Errorf("cannot scan user UUID: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "UPDATE users SET current_balance = current_balance + $2 WHERE id = $1", userUUID, order.Accrual); err != nil {
		return fmt.Errorf("cannot perform users update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("cannot finish transaction: %w", err)
	}

	return nil
}

func (r *repository) GetOrder(ctx context.Context, number string) (models.Order, error) {
	row := r.db.QueryRowContext(ctx, "SELECT * FROM orders WHERE number = $1", number)

	var order models.Order
	if err := row.Scan(&order); err != nil {
		return models.Order{}, fmt.Errorf("cannot parse row: %w", err)
	}

	return order, nil
}

func (r *repository) GetUser(ctx context.Context, uuid string) (models.User, error) {
	row := r.db.QueryRowxContext(ctx, "SELECT * FROM users WHERE id = $1", uuid)

	var user models.User
	if err := row.StructScan(&user); err != nil {
		return models.User{}, fmt.Errorf("cannot parse row: %w", err)
	}

	return user, nil
}

func (r *repository) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	row := r.db.QueryRowxContext(ctx, "SELECT * FROM users WHERE login = $1", login)

	var user models.User
	if err := row.StructScan(&user); errors.Is(err, sql.ErrNoRows) {
		return models.User{}, fmt.Errorf("%w: %w", ErrUserNotFound, err)
	} else if err != nil {
		return models.User{}, fmt.Errorf("cannot parse row: %w", err)
	}

	return user, nil
}

func (r *repository) ListOrders(ctx context.Context, uuid string) ([]models.Order, error) {
	rows, err := r.db.QueryxContext(ctx, "SELECT * FROM orders WHERE user_uuid = $1", uuid)
	if err != nil {
		return nil, fmt.Errorf("cannot query rows: %w", err)
	}

	var (
		orders []models.Order
		errs   []error
	)
	for rows.Next() {
		if err := rows.Err(); err != nil {
			errs = append(errs, err)
			continue
		}

		var order models.Order
		if err := rows.StructScan(&order); err != nil {
			errs = append(errs, err)
			continue
		}

		orders = append(orders, order)
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return orders, nil
}

func (r *repository) ListWithdrawals(ctx context.Context, uuid string) ([]models.Withdrawal, error) {
	rows, err := r.db.QueryxContext(ctx, "SELECT * FROM withdrawals WHERE user_uuid = $1", uuid)
	if err != nil {
		return nil, fmt.Errorf("cannot query rows: %w", err)
	}

	var (
		withdrawals []models.Withdrawal
		errs        []error
	)
	for rows.Next() {
		if err := rows.Err(); err != nil {
			errs = append(errs, err)
			continue
		}

		var withdrawal models.Withdrawal
		if err := rows.StructScan(&withdrawal); err != nil {
			errs = append(errs, err)
			continue
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return withdrawals, nil
}
