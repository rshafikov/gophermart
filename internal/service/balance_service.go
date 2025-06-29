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
var ErrInsufficientFunds = errors.New("insufficient funds")

type withdrawRepository interface {
	GetManyWithFilter(ctx context.Context, filter *models.WdFilter) ([]*models.Wd, error)
	CreateOne(ctx context.Context, tx *models.Wd) error
}

type balanceRepository interface {
	GetOneByUserID(ctx context.Context, userID int) (*models.Balance, error)
	UpdateOne(ctx context.Context, balance *models.Balance) error
}

type BalanceService struct {
	balanceRepo balanceRepository
	wdRepo      withdrawRepository
}

func NewBalanceService(balanceRepo balanceRepository, wdRepo withdrawRepository) *BalanceService {
	return &BalanceService{balanceRepo: balanceRepo, wdRepo: wdRepo}
}

func (s *BalanceService) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	userBalance, err := s.balanceRepo.GetOneByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFoundBalance
		}
	}

	queryFilter, err := repository.NewWdFilter(repository.WithUserID(userID))
	if err != nil {
		logger.L.Error("failed to build query filter", zap.Error(err))
		return nil, err
	}
	wds, err := s.wdRepo.GetManyWithFilter(ctx, queryFilter)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	var totalWithdrawal float64
	for _, withdrawal := range wds {
		totalWithdrawal += withdrawal.Amount
	}
	userBalance.Withdrawn = totalWithdrawal

	return userBalance, err
}

func (s *BalanceService) IncreaseUserBalance(ctx context.Context, userID int, amount float64) error {
	userBalance, err := s.balanceRepo.GetOneByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFoundBalance
		}
		return err
	}

	userBalance.Current += amount
	err = s.balanceRepo.UpdateOne(ctx, userBalance)
	if err != nil {
		logger.L.Error("unable to update balance", zap.Float64("new_balance", userBalance.Current))
		return err
	}
	return nil
}

func (s *BalanceService) GetWithdrawalsByUser(ctx context.Context, userID int) ([]*models.Wd, error) {
	filter, err := repository.NewWdFilter(repository.WithUserID(userID), repository.WithSorted())
	if err != nil {
		logger.L.Error("failed to build query filter", zap.Error(err))
		return nil, err
	}
	return s.wdRepo.GetManyWithFilter(ctx, filter)
}

func (s *BalanceService) Withdraw(ctx context.Context, b *models.Balance, wd *models.Wd) error {
	if b.Current < wd.Amount {
		logger.L.Error("insufficient funds",
			zap.Float64("current", b.Current),
			zap.Float64("requested", wd.Amount),
		)
		return ErrInsufficientFunds
	}

	b.Current -= wd.Amount
	wd.UserID = b.UserID
	wd.BalanceID = b.ID

	if err := s.balanceRepo.UpdateOne(ctx, b); err != nil {
		logger.L.Error("unable to update balance", zap.Float64("new_balance", b.Current))
		return err
	}

	if err := s.wdRepo.CreateOne(ctx, wd); err != nil {
		logger.L.Error("unable to create withdrawal",
			zap.Float64("amount", wd.Amount),
			zap.Float64("current_balance", b.Current),
		)
		return err
	}

	return nil
}
