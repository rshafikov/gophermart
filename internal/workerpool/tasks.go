package workerpool

import (
	"github.com/rshafikov/gophermart/internal/schemas"
)

type OrderTask struct {
	WorkerID          int
	OrderID           string
	UserID            int
	LastAccrualStatus schemas.AccrualStatus
	AccrualOrder      *schemas.AccrualOrder
	Error             error
}
