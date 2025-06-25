package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/models"
	"go.uber.org/zap"
)

var ErrOrderAlreadyLoaded = errors.New("order has already loaded")
var ErrOrderLoadedBySomeone = errors.New("order was loaded by another user")

type orderRepository interface {
	CreateOne(ctx context.Context, order *models.Order) error
	GetOneByNumeralID(ctx context.Context, numeralID string) (*models.Order, error)
	GetManyByUserID(ctx context.Context, userID int) ([]*models.Order, error)
	UpdateOne(ctx context.Context, order *models.Order) error
}

type OrderService struct {
	repo orderRepository
}

func NewOrderService(repo orderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrderIfNotExists(ctx context.Context, newOrder *models.Order) error {
	sameOrder, err := s.repo.GetOneByNumeralID(ctx, newOrder.NumeralID)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.repo.CreateOne(ctx, newOrder)
	}

	if sameOrder != nil && err == nil {
		logger.L.Error("order has been already loaded", zap.String("numeral_id", newOrder.NumeralID))
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

func (s *OrderService) UpdateOrder(ctx context.Context, order *models.Order) error {
	return s.repo.UpdateOne(ctx, order)
}
