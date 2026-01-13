// Package repository provides abstractions over DB storage.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darrior/gophermart/internal/models"
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
	db *sql.DB
}

func NewRepository(db *sql.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) AddUser(ctx context.Context, uuid, login, passHash string) error {
	if _, err := r.db.ExecContext(ctx, "INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)", uuid, login, passHash); err != nil {
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
	if _, err := tx.ExecContext(ctx, "UPDATE users SET balance = $1 WHERE id = $2", balance, uuid); err != nil {
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

func (r *repository) GetOrder(ctx context.Context, number string) (models.Order, error) {
	row := r.db.QueryRowContext(ctx, "SELECT * FROM orders WHERE number = $1", number)

	var order models.Order
	if err := row.Scan(&order); err != nil {
		return models.Order{}, fmt.Errorf("cannot parse row: %w", err)
	}

	return order, nil
}

func (r *repository) GetUser(ctx context.Context, uuid string) (models.User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", uuid)

	var user models.User
	if err := row.Scan(&user); err != nil {
		return models.User{}, fmt.Errorf("cannot parse row: %w", err)
	}

	return user, nil
}

func (r *repository) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT * FROM users WHERE login = $1", login)

	var user models.User
	if err := row.Scan(&user); err != nil {
		return models.User{}, fmt.Errorf("cannot parse row: %w", err)
	}

	return user, nil
}

func (r *repository) ListOrders(ctx context.Context, uuid string) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM orders WHERE user_uuid = $1", uuid)
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
		if err := rows.Scan(&order); err != nil {
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
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM withdrawals WHERE user_uuid = $1", uuid)
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
		if err := rows.Scan(&withdrawal); err != nil {
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
