package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appErr "github.com/Hlompy/Wallet/internal/errors"
	"github.com/gorilla/mux"
)

type MockWalletService struct {
	ProcessFunc func(ctx context.Context, walletID, op string, amount int64) error
	BalanceFunc func(ctx context.Context, walletID string) (int64, error)
}

func (m *MockWalletService) Process(ctx context.Context, walletID, op string, amount int64) error {
	if m.ProcessFunc != nil {
		return m.ProcessFunc(ctx, walletID, op, amount)
	}
	return nil
}

func (m *MockWalletService) Balance(ctx context.Context, walletID string) (int64, error) {
	if m.BalanceFunc != nil {
		return m.BalanceFunc(ctx, walletID)
	}
	return 0, nil
}

func TestPostWallet_Success(t *testing.T) {
	mockService := &MockWalletService{
		ProcessFunc: func(ctx context.Context, walletID, op string, amount int64) error {
			return nil
		},
		BalanceFunc: func(ctx context.Context, walletID string) (int64, error) {
			return 1000, nil
		},
	}

	handler := New(mockService)

	reqBody := walletRequest{
		WalletID: "11111111-1111-1111-1111-111111111123",
		OpType:   "DEPOSIT",
		Amount:   1000,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.PostWallet(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp walletResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.WalletID != reqBody.WalletID {
		t.Errorf("expected walletID %s, got %s", reqBody.WalletID, resp.WalletID)
	}
	if resp.Balance != 1000 {
		t.Errorf("expected balance 1000, got %d", resp.Balance)
	}
}

func TestPostWallet_InvalidJSON(t *testing.T) {
	handler := New(&MockWalletService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader([]byte("invalid")))
	rec := httptest.NewRecorder()

	handler.PostWallet(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPostWallet_InvalidWalletID(t *testing.T) {
	handler := New(&MockWalletService{})

	reqBody := walletRequest{
		WalletID: "not-a-uuid",
		OpType:   "DEPOSIT",
		Amount:   1000,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.PostWallet(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPostWallet_InsufficientFunds(t *testing.T) {
	mockService := &MockWalletService{
		ProcessFunc: func(ctx context.Context, walletID, op string, amount int64) error {
			return appErr.ErrInsufficientFunds
		},
	}

	handler := New(mockService)

	reqBody := walletRequest{
		WalletID: "550e8400-e29b-41d4-a716-446655440000",
		OpType:   "WITHDRAW",
		Amount:   1000,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.PostWallet(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestPostWallet_WalletNotFound(t *testing.T) {
	mockService := &MockWalletService{
		ProcessFunc: func(ctx context.Context, walletID, op string, amount int64) error {
			return appErr.ErrWalletNotFound
		},
	}

	handler := New(mockService)

	reqBody := walletRequest{
		WalletID: "550e8400-e29b-41d4-a716-446655440000",
		OpType:   "WITHDRAW",
		Amount:   1000,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.PostWallet(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestGetBalance_Success(t *testing.T) {
	mockService := &MockWalletService{
		BalanceFunc: func(ctx context.Context, walletID string) (int64, error) {
			return 5000, nil
		},
	}

	handler := New(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()

	// Имитируем mux.Vars
	req = mux.SetURLVars(req, map[string]string{
		"id": "550e8400-e29b-41d4-a716-446655440000",
	})

	handler.GetBalance(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]int64
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["balance"] != 5000 {
		t.Errorf("expected balance 5000, got %d", resp["balance"])
	}
}

func TestGetBalance_InvalidWalletID(t *testing.T) {
	handler := New(&MockWalletService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/invalid-id", nil)
	rec := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{
		"id": "invalid-id",
	})

	handler.GetBalance(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGetBalance_WalletNotFound(t *testing.T) {
	mockService := &MockWalletService{
		BalanceFunc: func(ctx context.Context, walletID string) (int64, error) {
			return 0, appErr.ErrWalletNotFound
		},
	}

	handler := New(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{
		"id": "550e8400-e29b-41d4-a716-446655440000",
	})

	handler.GetBalance(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
