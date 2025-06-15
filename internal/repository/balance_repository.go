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

type BalanceRepository struct {
	Pool *pgxpool.Pool
}

func NewBalanceRepository(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{Pool: pool}
}

func (r *BalanceRepository) GetOneByUserID(ctx context.Context, userID int) (*models.Balance, error) {
	var b models.Balance
	q := r.Pool.QueryRow(ctx, queries.GetBalanceByUserID, userID)
	err := q.Scan(&b.ID, &b.UserID, &b.Current, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.L.Error("there is no balance for user", zap.Int("user_id", userID))
			return nil, err
		}
		logger.L.Error("unable to get balance, unknown error", zap.Error(err))
		return nil, err
	}
	return &b, nil
}

func (r *BalanceRepository) UpdateOne(ctx context.Context, balance *models.Balance) error {
	return nil
}
