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

type TxRepository struct {
	Pool *pgxpool.Pool
}

type TxOption func(*models.TxFilter)

func NewTxFilter(options ...TxOption) (*models.TxFilter, error) {
	txFilter := models.TxFilter{}
	for _, applyOption := range options {
		applyOption(&txFilter)
	}

	if txFilter.Sorted {
		txFilter.FinalQuery = txFilter.Query[:len(txFilter.Query)-2] + queries.SortByCreatedAt
	} else {
		txFilter.FinalQuery = txFilter.Query
	}

	switch txFilter.Query {
	case queries.GetTxsByUserID:
		txFilter.EntityID = txFilter.UserID
	case queries.GetTxsByBalanceID:
		txFilter.EntityID = txFilter.BalanceID
	default:
		return nil, errors.New("invalid query filter")
	}

	return &txFilter, nil
}

func WithUserID(id int) TxOption {
	return func(filter *models.TxFilter) {
		filter.UserID = id
		filter.Query = queries.GetTxsByUserID
	}
}

func WithSorted() TxOption {
	return func(filter *models.TxFilter) {
		filter.Sorted = true
	}
}

func NewTxRepository(pool *pgxpool.Pool) *TxRepository {
	return &TxRepository{Pool: pool}
}

func (r *TxRepository) GetManyWithFilter(ctx context.Context, f *models.TxFilter) ([]*models.Tx, error) {
	q, err := r.Pool.Query(ctx, f.FinalQuery, f.EntityID)
	if err != nil {
		logger.L.Error("unable to get orders", zap.Error(err))
		return nil, err
	}
	defer q.Close()

	var txs []*models.Tx
	for q.Next() {
		var t models.Tx
		err = q.Scan(&t.ID, &t.UserID, &t.BalanceID, &t.OrderNumeralID, &t.Amount, &t.CreatedAt)
		if err != nil {
			logger.L.Error("failed to scan a tx", zap.Error(err))
			return nil, err
		}
		txs = append(txs, &t)
	}
	return txs, nil
}

//
//func (r *TxRepository) GetManyWithOrderNumeralID(ctx context.Context, userID int) ([]*models.Tx, error) {
//	q, err := r.Pool.Query(ctx, queries.GetTxsAndOrderNumberByUserID, userID)
//	if err != nil {
//		logger.L.Error("unable to get orders", zap.Error(err))
//		return nil, err
//	}
//	defer q.Close()
//
//	var txs []*models.Tx
//	for q.Next() {
//		var t models.Tx
//		err = q.Scan(&t.ID, &t.UserID, &t.BalanceID, &t.OrderID, &t.Amount, &t.CreatedAt, &t.OrderNumeralID)
//		if err != nil {
//			logger.L.Error("failed to scan a tx", zap.Error(err))
//			return nil, err
//		}
//		txs = append(txs, &t)
//	}
//	return txs, nil
//}
