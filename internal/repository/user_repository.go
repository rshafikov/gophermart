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
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		logger.L.Debug("error starting transaction", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	var userID int
	err = tx.QueryRow(ctx, queries.CreateUser, user.Login, user.Password).Scan(&userID)
	if err != nil {
		logger.L.Debug("unable to create user", zap.Error(err))
		return err
	}
	_, err = tx.Exec(ctx, queries.CreateBalance, userID)
	if err != nil {
		logger.L.Debug("unable to create balance for user", zap.Error(err), zap.Int("id", userID))
		return err
	}
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
