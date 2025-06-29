package schemas

type OrderCreateRequest struct {
	NumeralID string `json:"numeral_id"`
}

type OrderGetResponse struct {
	Accrual   float64 `json:"accrual,omitempty"`
	NumeralID string  `json:"number"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"uploaded_at"`
}

type AccrualStatus string

const (
	AccrualOrderStatusRegistered AccrualStatus = "REGISTERED"
	AccrualOrderStatusProcessing AccrualStatus = "PROCESSING"
	AccrualOrderStatusProcessed  AccrualStatus = "PROCESSED"
	AccrualOrderStatusInvalid    AccrualStatus = "INVALID"
)

type AccrualOrder struct {
	Accrual   float64       `json:"accrual,omitempty"`
	NumeralID string        `json:"number"`
	Status    AccrualStatus `json:"status"`
}
