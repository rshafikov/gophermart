package handlers

import (
	"encoding/json"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type BalanceHandler struct {
	BalanceService *service.BalanceService
}

func NewBalanceHandler(service *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{BalanceService: service}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userBalance, err := h.BalanceService.GetUserBalance(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("error getting user balance", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(userBalance)
	if err != nil {
		logger.L.Error("error marshalling user balance", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		logger.L.Error("failed to send balance in response", zap.Error(err))
	}

}

func (h *BalanceHandler) WithdrawFromBalance(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userBalance, err := h.BalanceService.GetUserBalance(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("error getting user balance", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *BalanceHandler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	txs, err := h.BalanceService.GetTxs(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("error getting user transactions", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	resp, err := json.Marshal(txs)
	if err != nil {
		logger.L.Error("error marshalling user transactions", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(txs) > 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	_, err = w.Write(resp)
	if err != nil {
		logger.L.Error("failed to send transactions in response", zap.Error(err))
	}
}
