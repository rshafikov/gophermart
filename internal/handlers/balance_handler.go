package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type balanceService interface {
	GetUserBalance(ctx context.Context, userID int) (*models.Balance, error)
	IncreaseUserBalance(ctx context.Context, userID int, amount float64) error
	Withdraw(ctx context.Context, balance *models.Balance, withdrawal *models.Wd) error
	GetWithdrawalsByUser(ctx context.Context, id int) ([]*models.Wd, error)
}

type BalanceHandler struct {
	BalanceService balanceService
}

func NewBalanceHandler(service balanceService) *BalanceHandler {
	return &BalanceHandler{BalanceService: service}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	userBalance, err := h.BalanceService.GetUserBalance(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("error getting user balance", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(userBalance)
	if err != nil {
		logger.L.Error("error marshalling user balance", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(resp)
	if err != nil {
		logger.L.Error("failed to send balance in response", zap.Error(err))
	}

}

func (h *BalanceHandler) WithdrawFromBalance(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	var withdraw models.Wd
	if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
		logger.L.Error("error decoding request body", zap.Error(err))
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if err := withdraw.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	userBalance, err := h.BalanceService.GetUserBalance(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("error getting user balance", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	if err := h.BalanceService.Withdraw(r.Context(), userBalance, &withdraw); err != nil {
		if errors.Is(err, service.ErrInsufficientFunds) {
			http.Error(w, "insufficient funds", http.StatusPaymentRequired)
			return
		}
		logger.L.Error("error updating balance", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(contextkeys.UserKey).(*models.User)
	if !ok {
		logger.L.Error("user not found in context")
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	wds, err := h.BalanceService.GetWithdrawalsByUser(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("error getting user withdrawals", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
	}

	resp, err := json.Marshal(wds)
	if err != nil {
		logger.L.Error("error marshalling user withdrawals", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(wds) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		logger.L.Error("failed to send withdrawals in response", zap.Error(err))
	}
}
