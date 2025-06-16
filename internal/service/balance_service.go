package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/repository"
	"go.uber.org/zap"
)

var ErrNotFoundBalance = errors.New("not found")

type BalanceService struct {
	balanceRepo models.BalanceRepository
	txRepo      models.TxRepository
}

func NewBalanceService(balanceRepo models.BalanceRepository, txRepo models.TxRepository) *BalanceService {
	return &BalanceService{balanceRepo: balanceRepo, txRepo: txRepo}
}

func (s *BalanceService) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	userBalance, err := s.balanceRepo.GetOneByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFoundBalance
		}
	}

	queryFilter, err := repository.NewTxFilter(repository.WithUserID(userID))
	if err != nil {
		logger.L.Error("failed to build query filter", zap.Error(err))
		return nil, err
	}
	txs, err := s.txRepo.GetManyWithFilter(ctx, queryFilter)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	var totalWithdrawal float64
	for _, transaction := range txs {
		totalWithdrawal += transaction.Amount
	}
	userBalance.Withdrawn = totalWithdrawal

	return userBalance, err
}

func (s *BalanceService) ChangeUserBalance(ctx context.Context, balance *models.Balance, tx *models.Tx) error {
	if err := s.txRepo.CreateOne(ctx, tx); err != nil {
		return err
	}

	balance.Current -= tx.Amount
	if err := s.balanceRepo.UpdateOne(ctx, balance); err != nil {
		logger.L.Error("unable to update user balance",
			zap.Float64("current_balance", balance.Current),
			zap.Float64("tx_amount", tx.Amount),
		)
		return err
	}

	return nil
}

func (s *BalanceService) GetTxs(ctx context.Context, userID int) ([]*models.Tx, error) {
	filter, err := repository.NewTxFilter(repository.WithUserID(userID), repository.WithSorted())
	if err != nil {
		logger.L.Error("failed to build query filter", zap.Error(err))
		return nil, err
	}
	return s.txRepo.GetManyWithFilter(ctx, filter)
}
