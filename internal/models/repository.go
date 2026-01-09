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
	UUID             string
	Login            string
	PasswordHash     string
	CurrentBalance   float64
	WithdrawnBalance float64
}

type Order struct {
	UserUUID   string
	Accural    int
	Number     string
	Status     OrderStatus
	UploadedAt time.Time
}

type Withdrawal struct {
	UserUUID    string
	Order       string
	Sum         float64
	ProcessedAt time.Time
}
