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

type OrderRepository struct {
	Pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{Pool: pool}
}

func (r *OrderRepository) CreateOne(ctx context.Context, o *models.Order) error {
	exec, err := r.Pool.Exec(ctx, queries.CreateOrder, o.NumeralID, o.UserID, o.Status, o.Accrual)
	if err != nil {
		logger.L.Error("unable to CREATE order", zap.Error(err))
		return err
	}
	logger.L.Debug("rows affected", zap.Int64("rows", exec.RowsAffected()))
	return nil
}

func (r *OrderRepository) GetOneByNumeralID(ctx context.Context, numeralID string) (*models.Order, error) {
	var o models.Order
	q := r.Pool.QueryRow(ctx, queries.GetOrderByNumeralID, numeralID)
	err := q.Scan(&o.ID, &o.NumeralID, &o.UserID, &o.Status, &o.Accrual, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.L.Error("there is no order with this number", zap.String("numeral_id", numeralID))
			return nil, err
		}
		logger.L.Error("unable to GET order, unknown error", zap.Error(err))
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) GetManyByUserID(ctx context.Context, userID int) ([]*models.Order, error) {
	q, err := r.Pool.Query(ctx, queries.GetOrdersByUserID, userID)
	if err != nil {
		logger.L.Error("unable to GET orders", zap.Error(err))
		return nil, err
	}
	defer q.Close()

	var orders []*models.Order
	for q.Next() {
		var o models.Order
		err = q.Scan(&o.ID, &o.NumeralID, &o.UserID, &o.Status, &o.Accrual, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			logger.L.Error("failed to scan order", zap.Error(err))
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}
