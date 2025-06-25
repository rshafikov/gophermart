package models

import (
	"time"
)

type Balance struct {
	ID        int       `json:"-"`
	UserID    int       `json:"-"`
	Current   float64   `json:"current"`
	Withdrawn float64   `json:"withdrawn"`
	UpdatedAt time.Time `json:"-"`
}
