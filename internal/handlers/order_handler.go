package handlers

import (
	"encoding/json"
	"errors"
	"github.com/rshafikov/gophermart/internal/client"
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

var ErrInvalidOrderNumber = errors.New("invalid order number")

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

	newOrder, err := h.validateOrder(body, u)
	if err != nil {
		logger.L.Error("invalid order number", zap.ByteString("body", body), zap.Error(err))
		http.Error(w, MsgInvalidOrderNumber, http.StatusUnprocessableEntity)
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
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}
}

func (h *OrderHandler) validateOrder(body []byte, user *models.User) (*models.Order, error) {
	numeralID := string(body)
	if numeralID == "" {
		return nil, errors.New("empty order number")
	}

	if _, err := strconv.Atoi(numeralID); err != nil {
		logger.L.Error(
			"unable to convert order number to a digit",
			zap.String("order number", numeralID),
			zap.Error(err),
		)
		return nil, ErrInvalidOrderNumber
	}

	if isNumeralIDValid := security.LuhnAlgoPredicat(numeralID); !isNumeralIDValid {
		logger.L.Error("order number doesn't pass Luhn algorithm", zap.String("order number", numeralID))
		return nil, ErrInvalidOrderNumber
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
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	orders, err := h.Service.GetOrders(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("failed to get orders", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
	}

	resp, err := json.Marshal(orders)
	if err != nil {
		logger.L.Error("failed to marshal orders", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		logger.L.Error("failed to send orders in response", zap.Error(err))
	}
}
