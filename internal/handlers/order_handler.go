package handlers

import (
	"encoding/json"
	"errors"
	"github.com/rshafikov/gophermart/internal/client"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/service"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const MsgInternalServerError = "internal server error"
const MsgInvalidOrderNumber = "invalid order number"

type OrderHandler struct {
	Service *service.OrderService
	Client  client.Client
}

func NewOrderHandler(orderService *service.OrderService, client client.Client) *OrderHandler {
	return &OrderHandler{Service: orderService, Client: client}
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

	newOrder := models.Order{NumeralID: string(body), UserID: u.ID}

	if err = newOrder.Validate(); err != nil {
		logger.L.Error("invalid order number", zap.ByteString("body", body), zap.Error(err))
		http.Error(w, MsgInvalidOrderNumber, http.StatusUnprocessableEntity)
		return
	}

	externalOrder, err := h.Client.GetOrderStatus(r.Context(), newOrder.NumeralID)
	if err != nil {
		logger.L.Error("failed to get order status", zap.Error(err), zap.String("numeral_id", newOrder.NumeralID))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}
	newOrder.Accrual = externalOrder.Accrual
	newOrder.Status = models.OrderStatus(externalOrder.Status)

	err = h.Service.CreateOrderIfNotExists(r.Context(), &newOrder)
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
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
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
