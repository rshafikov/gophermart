package models

import (
	"context"
	"errors"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"go.uber.org/zap"
	"time"
)

var ErrWdOrderIDValidationFailed = errors.New("invalid order number format")
var ErrWdAmountValidationFailed = errors.New("invalid amount to withdraw")

type Wd struct {
	ID             int       `json:"-"`
	UserID         int       `json:"-"`
	BalanceID      int       `json:"-"`
	OrderNumeralID string    `json:"order"`
	Amount         float64   `json:"sum"`
	CreatedAt      time.Time `json:"processed_at,omitempty"`
}

func (w *Wd) Validate() error {
	if !security.LuhnAlgoPredicat(w.OrderNumeralID) {
		logger.L.Error("invalid order number format", zap.String("order_numeral_id", w.OrderNumeralID))
		return ErrWdOrderIDValidationFailed
	}

	if w.Amount <= 0 {
		logger.L.Error("invalid amount to withdraw", zap.Float64("amount", w.Amount))
		return ErrWdAmountValidationFailed
	}

	return nil
}

type WdFilter struct {
	BalanceID  int
	UserID     int
	Query      string
	FinalQuery string
	Sorted     bool
	EntityID   int
}

type WdRepository interface {
	GetManyWithFilter(ctx context.Context, filter *WdFilter) ([]*Wd, error)
	CreateOne(ctx context.Context, tx *Wd) error
}
