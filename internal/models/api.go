package models

type AuthenticationData struct {
	Login    string `json:"string"`
	Password string `json:"password"`
}

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accural"`
	UploadedAt string `json:"uploaded_at"`
}

type Balance struct {
	Current   string `json:"current"`
	Withdrawn string `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string `json:"string"`
	Sum   int    `json:"sum"`
}

type Withdraw struct {
	Order       string `json:"string"`
	Sum         int    `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}
