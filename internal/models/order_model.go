package models

import (
	"errors"
	"fmt"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type InternalOrderStatus string

const (
	OrderStatusNew        InternalOrderStatus = "NEW"
	OrderStatusProcessing InternalOrderStatus = "PROCESSING"
	OrderStatusInvalid    InternalOrderStatus = "INVALID"
	OrderStatusProcessed  InternalOrderStatus = "PROCESSED"
)

type Order struct {
	ID        int                 `json:"-"`
	NumeralID string              `json:"number"`
	UserID    int                 `json:"-"`
	Status    InternalOrderStatus `json:"status"`
	Accrual   float64             `json:"accrual,omitempty"`
	CreatedAt time.Time           `json:"uploaded_at"`
	UpdatedAt time.Time           `json:"-"`
}

func (o *Order) Validate() error {
	if o.NumeralID == "" {
		return errors.New("empty order number")
	}

	if _, err := strconv.Atoi(o.NumeralID); err != nil {
		logger.L.Error(
			"unable to convert order number to a digit",
			zap.String("order number", o.NumeralID),
			zap.Error(err),
		)
		return fmt.Errorf("invalid order number: %s", o.NumeralID)
	}

	if isNumeralIDValid := security.LuhnAlgoPredicat(o.NumeralID); !isNumeralIDValid {
		logger.L.Error("order number doesn't pass Luhn algorithm", zap.String("order number", o.NumeralID))
		return fmt.Errorf("invalid order number: %s", o.NumeralID)
	}

	return nil
}
