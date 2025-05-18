package service

import (
	"context"
	"errors"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/models"
	"log"
)

var ErrPasswordMismatch = errors.New("password mismatch")
var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("login is not available")
var ErrDB = errors.New("database error")

type UserService struct {
	repo models.UserRepository
}

func NewUserService(repo models.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, login, password string) error {
	oldUser, _ := s.repo.GetByLogin(ctx, login)
	if oldUser != nil {
		return ErrUserAlreadyExists
	}

	password, err := security.HashPassword(password)
	if err != nil {
		return ErrDB
	}

	err = s.repo.CreateUser(ctx, &models.User{Login: login, Password: password})
	if err != nil {
		return ErrDB
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, login, password string) (*models.User, error) {
	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		log.Println("unable to GET user by login:", err)
		return nil, ErrUserNotFound
	}

	checkPassword := security.CheckPasswordHash(password, user.Password)
	if !checkPassword {
		return nil, ErrPasswordMismatch
	}

	return user, nil
}

func (s *UserService) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		log.Println("unable to GET user by login:", err)
		return nil, ErrUserNotFound
	}

	return user, nil
}
