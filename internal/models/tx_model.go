package models

import (
	"context"
	"time"
)

type Tx struct {
	ID             int       `json:"-"`
	UserID         int       `json:"-"`
	BalanceID      int       `json:"-"`
	OrderNumeralID string    `json:"order"`
	Amount         float64   `json:"sum"`
	CreatedAt      time.Time `json:"processed_at"`
}

type TxFilter struct {
	BalanceID  int
	UserID     int
	Query      string
	FinalQuery string
	Sorted     bool
	EntityID   int
}

type TxRepository interface {
	GetManyWithFilter(ctx context.Context, filter *TxFilter) ([]*Tx, error)
	CreateOne(ctx context.Context, tx *Tx) error
}
