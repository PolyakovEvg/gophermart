package handler

import (
	"encoding/json"
	"errors"
	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/service"
	"net/http"
	"time"

	"go-musthave-diploma-tpl/internal/repository/postgres"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WithdrawalResponse struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt string          `json:"processed_at"`
}

type BalanceHandler struct {
	service service.BalanceServicer
	logger  *zap.Logger
}

func NewBalanceHandler(s service.BalanceServicer, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{service: s, logger: logger}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	current, withdrawn, err := h.service.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get balance", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := map[string]decimal.Decimal{
		"current":   current,
		"withdrawn": withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Order string          `json:"order"`
		Sum   decimal.Decimal `json:"sum"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode withdraw request", zap.Error(err))
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Sum.LessThanOrEqual(decimal.Zero) {
		http.Error(w, "sum must be greater than zero", http.StatusUnprocessableEntity)
		return
	}

	err := h.service.Withdraw(r.Context(), userID, req.Order, req.Sum)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, postgres.ErrNotEnoughFunds):
		http.Error(w, "insufficient funds", http.StatusPaymentRequired)
	case errors.Is(err, postgres.ErrInvalidOrder):
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
	default:
		h.logger.Error("withdraw error", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *BalanceHandler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	list, err := h.service.ListWithdrawals(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list withdrawals", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]WithdrawalResponse, 0, len(list))
	for _, wdr := range list {
		resp = append(resp, WithdrawalResponse{
			Order:       wdr.OrderNumber,
			Sum:         wdr.Sum,
			ProcessedAt: wdr.ProcessedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}
