package models

import (
	"context"
	"time"
)

type OrderStatus string

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	ID        int         `json:"-"`
	NumeralID string      `json:"number"`
	UserID    int         `json:"-"`
	Status    OrderStatus `json:"status"`
	Accrual   int         `json:"accrual,omitempty"`
	CreatedAt time.Time   `json:"uploaded_at"`
	UpdatedAt time.Time   `json:"-"`
}

type OrderRepository interface {
	CreateOne(ctx context.Context, order *Order) error
	GetOneByNumeralID(ctx context.Context, numeralID string) (*Order, error)
	GetManyByUserID(ctx context.Context, userID int) ([]*Order, error)
}

type OrderService interface {
	CreateOrderIfNotExists(ctx context.Context, order *Order) error
	GetOrders(ctx context.Context, userID int) ([]*Order, error)
}
