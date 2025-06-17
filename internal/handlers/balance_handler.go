package handlers

import (
	"encoding/json"
	"github.com/rshafikov/gophermart/internal/core/contextkeys"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/schemas"
	"github.com/rshafikov/gophermart/internal/service"
	"go.uber.org/zap"
	"net/http"
	"time"
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

	var withdraw schemas.Withdraw
	if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
		logger.L.Error("error decoding request body", zap.Error(err))
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if !security.LuhnAlgoPredicat(withdraw.Order) {
		logger.L.Error("invalid order number format")
		http.Error(w, MsgInvalidOrderNumber, http.StatusUnprocessableEntity)
		return
	}

	if withdraw.Sum <= 0 {
		logger.L.Error("invalid sum value")
		http.Error(w, "invalid sum value", http.StatusUnprocessableEntity)
		return
	}

	userBalance, err := h.BalanceService.GetUserBalance(r.Context(), u.ID)
	if err != nil {
		logger.L.Error("error getting user balance", zap.Error(err))
		http.Error(w, MsgInternalServerError, http.StatusInternalServerError)
		return
	}

	if userBalance.Current < withdraw.Sum {
		logger.L.Error("insufficient funds",
			zap.Float64("current", userBalance.Current),
			zap.Float64("requested", withdraw.Sum),
		)
		http.Error(w, "insufficient funds", http.StatusPaymentRequired)
		return
	}

	wd := &models.Wd{
		UserID:         u.ID,
		BalanceID:      userBalance.ID,
		OrderNumeralID: withdraw.Order,
		Amount:         withdraw.Sum,
		CreatedAt:      time.Now(),
	}

	userBalance.Current -= withdraw.Sum
	if err := h.BalanceService.ChangeUserBalance(r.Context(), userBalance, wd); err != nil {
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
