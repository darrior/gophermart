package models

import "time"

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type User struct {
	UUID             string  `db:"id"`
	Login            string  `db:"login"`
	PasswordHash     string  `db:"password_hash"`
	CurrentBalance   float64 `db:"current_balance"`
	WithdrawnBalance float64 `db:"withdrawan_balance"`
}

type Order struct {
	UserUUID   string      `db:"user_uuid"`
	Accrual    float64     `db:"accrual"`
	Number     string      `db:"number"`
	Status     OrderStatus `db:"status"`
	UploadedAt time.Time   `db:"uploaded_at"`
}

type Withdrawal struct {
	ID          int       `db:"id"`
	UserUUID    string    `db:"user_uuid"`
	Order       string    `db:"order"`
	Sum         float64   `db:"sum"`
	ProcessedAt time.Time `db:"processed_at"`
}
