package models

import (
	"context"
	"time"
)

type Wd struct {
	ID             int       `json:"-"`
	UserID         int       `json:"-"`
	BalanceID      int       `json:"-"`
	OrderNumeralID string    `json:"order"`
	Amount         float64   `json:"sum"`
	CreatedAt      time.Time `json:"processed_at"`
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
