package models

import (
	"context"
	"time"
)

type User struct {
	ID        int
	Login     string
	Password  string
	CreatedAt time.Time
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetByLogin(ctx context.Context, login string) (*User, error)
}

type UserService interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) (*User, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
}
