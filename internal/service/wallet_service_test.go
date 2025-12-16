package service

import (
	"context"
	"testing"

	appErr "github.com/Hlompy/Wallet/internal/errors"
)

type MockWalletRepository struct {
	UpdateBalanceFunc func(ctx context.Context, walletID string, amount int64) error
	GetBalanceFunc    func(ctx context.Context, walletID string) (int64, error)
}

func (m *MockWalletRepository) UpdateBalance(ctx context.Context, walletID string, amount int64) error {
	if m.UpdateBalanceFunc != nil {
		return m.UpdateBalanceFunc(ctx, walletID, amount)
	}
	return nil
}

func (m *MockWalletRepository) GetBalance(ctx context.Context, walletID string) (int64, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, walletID)
	}
	return 0, nil
}

func TestProcess_Deposit(t *testing.T) {
	mockRepo := &MockWalletRepository{
		UpdateBalanceFunc: func(ctx context.Context, walletID string, amount int64) error {
			if amount != 1000 {
				t.Errorf("expected amount 1000, got %d", amount)
			}
			return nil
		},
	}

	service := New(mockRepo)

	err := service.Process(context.Background(), "test-wallet", "DEPOSIT", 1000)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcess_Withdraw(t *testing.T) {
	mockRepo := &MockWalletRepository{
		UpdateBalanceFunc: func(ctx context.Context, walletID string, amount int64) error {
			if amount != -500 {
				t.Errorf("expected amount -500, got %d", amount)
			}
			return nil
		},
	}

	service := New(mockRepo)

	err := service.Process(context.Background(), "test-wallet", "WITHDRAW", 500)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProcess_InvalidOperation(t *testing.T) {
	service := New(&MockWalletRepository{})

	tests := []struct {
		name   string
		op     string
		amount int64
	}{
		{"invalid operation type", "INVALID", 100},
		{"zero amount", "DEPOSIT", 0},
		{"negative amount", "DEPOSIT", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Process(context.Background(), "test-wallet", tt.op, tt.amount)
			if err != appErr.ErrInvalidOperation {
				t.Errorf("expected ErrInvalidOperation, got %v", err)
			}
		})
	}
}

func TestProcess_RepositoryError(t *testing.T) {
	mockRepo := &MockWalletRepository{
		UpdateBalanceFunc: func(ctx context.Context, walletID string, amount int64) error {
			return appErr.ErrInsufficientFunds
		},
	}

	service := New(mockRepo)

	err := service.Process(context.Background(), "test-wallet", "WITHDRAW", 1000)
	if err != appErr.ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestBalance_Success(t *testing.T) {
	mockRepo := &MockWalletRepository{
		GetBalanceFunc: func(ctx context.Context, walletID string) (int64, error) {
			return 5000, nil
		},
	}

	service := New(mockRepo)

	balance, err := service.Balance(context.Background(), "test-wallet")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if balance != 5000 {
		t.Errorf("expected balance 5000, got %d", balance)
	}
}

func TestBalance_WalletNotFound(t *testing.T) {
	mockRepo := &MockWalletRepository{
		GetBalanceFunc: func(ctx context.Context, walletID string) (int64, error) {
			return 0, appErr.ErrWalletNotFound
		},
	}

	service := New(mockRepo)

	balance, err := service.Balance(context.Background(), "test-wallet")
	if err != appErr.ErrWalletNotFound {
		t.Errorf("expected ErrWalletNotFound, got %v", err)
	}

	if balance != 0 {
		t.Errorf("expected balance 0, got %d", balance)
	}
}

func TestBalance_RepositoryError(t *testing.T) {
	expectedErr := appErr.ErrWalletNotFound

	mockRepo := &MockWalletRepository{
		GetBalanceFunc: func(ctx context.Context, walletID string) (int64, error) {
			return 0, expectedErr
		},
	}

	service := New(mockRepo)

	_, err := service.Balance(context.Background(), "test-wallet")
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
