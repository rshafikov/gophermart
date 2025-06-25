package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/service"
	"github.com/rshafikov/gophermart/internal/workerpool"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const MsgInternalServerError = "internal server error"
const MsgInvalidOrderNumber = "invalid order number"

type orderService interface {
	CreateOrderIfNotExists(ctx context.Context, order *models.Order) error
	GetOrders(ctx context.Context, userID int) ([]*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
}

type OrderHandler struct {
	Service orderService
	WP      *workerpool.WorkerPool
}

func NewOrderHandler(orderService orderService, pool *workerpool.WorkerPool) *OrderHandler {
	return &OrderHandler{Service: orderService, WP: pool}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	newOrder := models.Order{NumeralID: string(body), UserID: u.ID, Status: models.OrderStatusNew}

	if err = newOrder.Validate(); err != nil {
		logger.L.Error("invalid order number", zap.ByteString("body", body), zap.Error(err))
		http.Error(w, MsgInvalidOrderNumber, http.StatusUnprocessableEntity)
		return
	}

	err = h.Service.CreateOrderIfNotExists(r.Context(), &newOrder)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
		h.WP.AddTask(workerpool.OrderTask{
			OrderID:           newOrder.NumeralID,
			UserID:            u.ID,
			LastAccrualStatus: "",
			AccrualOrder:      nil,
			Error:             nil,
		})
	case errors.Is(err, service.ErrOrderAlreadyLoaded):
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrOrderLoadedBySomeone):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		logger.L.Error("failed to create order", zap.Error(err), zap.String("numeral_id", newOrder.NumeralID))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
	}

}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	orders, err := h.Service.GetOrders(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("failed to get orders", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(orders)
	if err != nil {
		logger.L.Error("failed to marshal orders", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(resp)
	if err != nil {
		logger.L.Error("failed to send orders in response", zap.Error(err))
	}
}
