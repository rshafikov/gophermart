package models

import (
	"context"
	"time"
)

type Balance struct {
	ID        int       `json:"-"`
	UserID    int       `json:"-"`
	Current   float64   `json:"current"`
	Withdrawn float64   `json:"withdrawn"`
	UpdatedAt time.Time `json:"-"`
}

type BalanceRepository interface {
	GetOneByUserID(ctx context.Context, userID int) (*Balance, error)
	UpdateOne(ctx context.Context, balance *Balance) error
}

type BalanceService interface {
	GetUserBalance(ctx context.Context, userID int) (*Balance, error)
	ChangeUserBalance(ctx context.Context, balance *Balance, tx *Tx) error
}
