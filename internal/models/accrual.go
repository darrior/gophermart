package models

const (
	AccrualOrderStatusRegistered = "REGISTERED"
	AccrualOrderStatusInvalid    = "INVALID"
	AccrualOrderStatusProcessing = "PROCESSING"
	AccrualOrderStatusProcessed  = "PROCESSED"
)

type AccrualOrderStatus string

type AccrualOrderState struct {
	Order   string             `json:"order"`
	Status  AccrualOrderStatus `json:"status"`
	Accrual float64            `json:"accrual"`
}
