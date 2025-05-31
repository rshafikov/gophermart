package handlers

import (
	"encoding/json"
	"errors"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/service"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

type OrderHandler struct {
	Service *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{Service: orderService}
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

	newOrder, err := h.validateOrder(body, u)
	if err != nil {
		logger.L.Error("invalid order number", zap.ByteString("body", body), zap.Error(err))
		http.Error(w, `{"error": "invalid order number"}`, http.StatusUnprocessableEntity)
		return
	}

	err = h.Service.CreateOrderIfNotExists(r.Context(), newOrder)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
		return
	case errors.Is(err, service.ErrOrderAlreadyLoaded):
		w.WriteHeader(http.StatusOK)
		return
	case errors.Is(err, service.ErrOrderLoadedBySomeone):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	default:
		logger.L.Error("failed to create order", zap.Error(err), zap.String("numeral_id", newOrder.NumeralID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *OrderHandler) validateOrder(body []byte, user *models.User) (*models.Order, error) {
	numeralID := string(body)
	if numeralID == "" {
		return nil, errors.New("empty order number")
	}

	if _, err := strconv.Atoi(numeralID); err != nil {
		return nil, errors.New("invalid order number")
	}

	if isNumeralIDValid := security.LuhnAlgoPredicat(numeralID); !isNumeralIDValid {
		return nil, errors.New("order number doesn't pass Luhn algorithm")
	}

	return &models.Order{
		NumeralID: string(body),
		UserID:    user.ID,
		Status:    models.StatusNew,
		Accrual:   0,
	}, nil
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	orders, err := h.Service.GetOrders(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("failed to get orders", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	resp, err := json.Marshal(orders)
	if err != nil {
		logger.L.Error("failed to marshal orders", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		logger.L.Error("failed to write orders", zap.Error(err))
	}
}
