package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/database/queries"
	"github.com/rshafikov/gophermart/internal/models"
	"go.uber.org/zap"
)

type UserRepository struct {
	Pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{Pool: pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	exec, err := r.Pool.Exec(ctx, queries.CreateUser, user.Login, user.Password)
	if err != nil {
		logger.L.Debug("unable to CREATE user", zap.Error(err))
		return err
	}
	logger.L.Debug("rows affected", zap.Int64("rows", exec.RowsAffected()))
	return nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User

	q := r.Pool.QueryRow(ctx, queries.GetUserByLogin, login)
	err := q.Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.L.Debug("there is no user with this login", zap.String("login", login))
			return nil, err
		}
		logger.L.Debug("unable to GET user, unknown error", zap.Error(err))
		return nil, err
	}
	return &user, nil
}
