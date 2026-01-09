package models

type AuthenticationDataRequest struct {
	Login    string `json:"string"`
	Password string `json:"password"`
}

type OrderResponse struct {
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    int         `json:"accural"`
	UploadedAt string      `json:"uploaded_at"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"string"`
	Sum   float64 `json:"sum"`
}

type WithdrawResponse struct {
	Order       string  `json:"string"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
