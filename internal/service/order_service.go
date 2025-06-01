package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/models"
	"go.uber.org/zap"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrOrderAlreadyLoaded = errors.New("order has already loaded")
var ErrOrderLoadedBySomeone = errors.New("order was loaded by another user")

type OrderService struct {
	repo models.OrderRepository
}

func NewOrderService(repo models.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrderIfNotExists(ctx context.Context, newOrder *models.Order) error {
	sameOrder, err := s.repo.GetOneByNumeralID(ctx, newOrder.NumeralID)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.repo.CreateOne(ctx, newOrder)
	}

	if sameOrder != nil && err == nil {
		logger.L.Debug("order has been already loaded", zap.String("numeral_id", newOrder.NumeralID))
		if sameOrder.UserID == newOrder.UserID {
			return ErrOrderAlreadyLoaded
		}
		return ErrOrderLoadedBySomeone
	}

	return err
}

func (s *OrderService) GetOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	return s.repo.GetManyByUserID(ctx, userID)
}
