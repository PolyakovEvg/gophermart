package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-musthave-diploma-tpl/internal/handler"
	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type mockBalanceService struct {
	GetBalanceFunc      func(ctx context.Context, userID string) (decimal.Decimal, decimal.Decimal, error)
	WithdrawFunc        func(ctx context.Context, userID, order string, sum decimal.Decimal) error
	ListWithdrawalsFunc func(ctx context.Context, userID string) ([]postgres.Withdrawal, error)
}

func (m *mockBalanceService) GetBalance(ctx context.Context, userID string) (decimal.Decimal, decimal.Decimal, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, userID)
	}
	return decimal.Zero, decimal.Zero, nil
}

func (m *mockBalanceService) Withdraw(ctx context.Context, userID, order string, sum decimal.Decimal) error {
	if m.WithdrawFunc != nil {
		return m.WithdrawFunc(ctx, userID, order, sum)
	}
	return nil
}

func (m *mockBalanceService) ListWithdrawals(ctx context.Context, userID string) ([]postgres.Withdrawal, error) {
	if m.ListWithdrawalsFunc != nil {
		return m.ListWithdrawalsFunc(ctx, userID)
	}
	return nil, nil
}

func TestBalanceHandler_GetBalance(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockBalanceService{
		GetBalanceFunc: func(ctx context.Context, userID string) (decimal.Decimal, decimal.Decimal, error) {
			withdrawn := decimal.NewFromInt(200)
			accrued := decimal.NewFromInt(1200)
			current := accrued.Sub(withdrawn)
			return current, withdrawn, nil
		},
	}

	h := handler.NewBalanceHandler(mockSvc, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, "user1"))
	rr := httptest.NewRecorder()
	h.GetBalance(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]decimal.Decimal
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedCurrent := decimal.NewFromInt(1000)
	expectedWithdrawn := decimal.NewFromInt(200)

	if !resp["current"].Equal(expectedCurrent) {
		t.Errorf("current: got %s, want %s", resp["current"].String(), expectedCurrent.String())
	}
	if !resp["withdrawn"].Equal(expectedWithdrawn) {
		t.Errorf("withdrawn: got %s, want %s", resp["withdrawn"].String(), expectedWithdrawn.String())
	}
}

func TestBalanceHandler_Withdraw(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockBalanceService{
		WithdrawFunc: func(ctx context.Context, userID, order string, sum decimal.Decimal) error {
			if sum.LessThanOrEqual(decimal.Zero) {
				return postgres.ErrInvalidOrder
			}
			if sum.GreaterThan(decimal.NewFromInt(1000)) {
				return postgres.ErrNotEnoughFunds
			}
			return nil
		},
	}

	h := handler.NewBalanceHandler(mockSvc, logger)

	tests := []struct {
		name           string
		userID         string
		body           string
		wantStatusCode int
	}{
		{"success", "user1", `{"order":"123","sum":100}`, http.StatusOK},
		{"invalid order (zero sum)", "user1", `{"order":"123","sum":0}`, http.StatusUnprocessableEntity},
		{"insufficient funds", "user1", `{"order":"123","sum":1500}`, http.StatusPaymentRequired},
		{"no user", "", `{"order":"123","sum":100}`, http.StatusUnauthorized},
		{"bad body", "user1", `{`, http.StatusBadRequest},
		{"invalid sum format", "user1", `{"order":"123","sum":"abc"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(tt.body))
			if tt.userID != "" {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, tt.userID))
			}
			rr := httptest.NewRecorder()
			h.Withdraw(rr, req)
			if rr.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestBalanceHandler_ListWithdrawals(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockBalanceService{
		ListWithdrawalsFunc: func(ctx context.Context, userID string) ([]postgres.Withdrawal, error) {
			if userID == "empty" {
				return []postgres.Withdrawal{}, nil
			}
			if userID == "error" {
				return nil, errors.New("database error")
			}
			return []postgres.Withdrawal{
				{
					OrderNumber: "123",
					Sum:         decimal.NewFromFloat(100.50),
					ProcessedAt: time.Now(),
				},
				{
					OrderNumber: "456",
					Sum:         decimal.NewFromFloat(200.75),
					ProcessedAt: time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewBalanceHandler(mockSvc, logger)

	tests := []struct {
		name           string
		userID         string
		wantStatusCode int
		wantBodyCount  int
	}{
		{"success", "user1", http.StatusOK, 2},
		{"no withdrawals", "empty", http.StatusNoContent, 0},
		{"unauthorized", "", http.StatusUnauthorized, 0},
		{"database error", "error", http.StatusInternalServerError, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/balance/withdrawals", nil)
			if tt.userID != "" {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, tt.userID))
			}
			rr := httptest.NewRecorder()
			h.ListWithdrawals(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatusCode)
			}

			if tt.wantBodyCount > 0 {
				var resp []handler.WithdrawalResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if len(resp) != tt.wantBodyCount {
					t.Errorf("got %d items, want %d", len(resp), tt.wantBodyCount)
				}

				for i, item := range resp {
					if item.Order == "" {
						t.Errorf("item %d: empty order number", i)
					}
				}
			}
		})
	}
}

func TestBalanceHandler_WithdrawalsResponseFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockBalanceService{
		ListWithdrawalsFunc: func(ctx context.Context, userID string) ([]postgres.Withdrawal, error) {
			return []postgres.Withdrawal{
				{
					OrderNumber: "123",
					Sum:         decimal.NewFromFloat(100.50),
					ProcessedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	h := handler.NewBalanceHandler(mockSvc, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance/withdrawals", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, "user1"))
	rr := httptest.NewRecorder()

	h.ListWithdrawals(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status OK, got %d", rr.Code)
	}

	var resp []handler.WithdrawalResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("expected 1 withdrawal, got %d", len(resp))
	}

	if resp[0].ProcessedAt != "2024-01-01T12:00:00Z" {
		t.Errorf("wrong date format: got %s, want 2024-01-01T12:00:00Z", resp[0].ProcessedAt)
	}

	t.Logf("Withdrawal response: %+v", resp[0])
}
