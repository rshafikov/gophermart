package schemas

type OrderCreateRequest struct {
	NumeralID string `json:"numeral_id"`
}

type OrderGetResponse struct {
	NumeralID string  `json:"number"`
	Status    string  `json:"status"`
	Accrual   float32 `json:"accrual,omitempty"`
	CreatedAt string  `json:"uploaded_at"`
}

type AccrualOrder struct {
	NumeralID string  `json:"number"`
	Status    string  `json:"status"`
	Accrual   float32 `json:"accrual"`
}
