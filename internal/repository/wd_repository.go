package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/database/queries"
	"github.com/rshafikov/gophermart/internal/models"
	"go.uber.org/zap"
)

type WdRepository struct {
	Pool *pgxpool.Pool
}

type WdOption func(*models.WdFilter)

func NewWdFilter(options ...WdOption) (*models.WdFilter, error) {
	wdFilter := models.WdFilter{}
	for _, applyOption := range options {
		applyOption(&wdFilter)
	}

	if wdFilter.Sorted {
		wdFilter.FinalQuery = wdFilter.Query[:len(wdFilter.Query)-2] + queries.SortByCreatedAt
	} else {
		wdFilter.FinalQuery = wdFilter.Query
	}

	switch wdFilter.Query {
	case queries.GetWdsByUserID:
		wdFilter.EntityID = wdFilter.UserID
	case queries.GetWdsByBalanceID:
		wdFilter.EntityID = wdFilter.BalanceID
	default:
		return nil, errors.New("invalid query filter")
	}

	return &wdFilter, nil
}

func WithUserID(id int) WdOption {
	return func(filter *models.WdFilter) {
		filter.UserID = id
		filter.Query = queries.GetWdsByUserID
	}
}

func WithSorted() WdOption {
	return func(filter *models.WdFilter) {
		filter.Sorted = true
	}
}

func NewWdRepository(pool *pgxpool.Pool) *WdRepository {
	return &WdRepository{Pool: pool}
}

func (r *WdRepository) GetManyWithFilter(ctx context.Context, f *models.WdFilter) ([]*models.Wd, error) {
	q, err := r.Pool.Query(ctx, f.FinalQuery, f.EntityID)
	if err != nil {
		logger.L.Error("unable to get withdrawals", zap.Error(err))
		return nil, err
	}
	defer q.Close()

	var wds []*models.Wd
	for q.Next() {
		var w models.Wd
		err = q.Scan(&w.ID, &w.UserID, &w.BalanceID, &w.OrderNumeralID, &w.Amount, &w.CreatedAt)
		if err != nil {
			logger.L.Error("failed to scan a withdrawal", zap.Error(err))
			return nil, err
		}
		wds = append(wds, &w)
	}
	return wds, nil
}

func (r *WdRepository) CreateOne(ctx context.Context, wd *models.Wd) error {
	_, err := r.Pool.Exec(ctx, queries.CreateWd, wd.UserID, wd.BalanceID, wd.OrderNumeralID, wd.Amount, wd.CreatedAt)
	if err != nil {
		logger.L.Error("unable to create withdrawal", zap.Reflect("wd", wd), zap.Error(err))
		return err
	}
	return nil
}
