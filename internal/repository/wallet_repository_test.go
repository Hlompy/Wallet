package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	appErr "github.com/Hlompy/Wallet/internal/errors"
)

func TestUpdateBalance_CreateWallet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	walletID := "550e8400-e29b-41d4-a716-446655440000"
	amount := int64(1000)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO wallets \(id, balance\) VALUES \(\$1, \$2\)`).
		WithArgs(walletID, amount).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.UpdateBalance(context.Background(), walletID, amount)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateBalance_CreateWallet_WithNegativeAmount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	walletID := "550e8400-e29b-41d4-a716-446655440000"
	amount := int64(-500)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err = repo.UpdateBalance(context.Background(), walletID, amount)
	if err != appErr.ErrWalletNotFound {
		t.Errorf("expected ErrWalletNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateBalance_Deposit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	walletID := "550e8400-e29b-41d4-a716-446655440000"
	currentBalance := int64(1000)
	amount := int64(500)
	newBalance := currentBalance + amount

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(currentBalance))
	mock.ExpectExec(`UPDATE wallets SET balance = \$1 WHERE id = \$2`).
		WithArgs(newBalance, walletID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.UpdateBalance(context.Background(), walletID, amount)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateBalance_Withdraw_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	walletID := "550e8400-e29b-41d4-a716-446655440000"
	currentBalance := int64(1000)
	amount := int64(-500)
	newBalance := currentBalance + amount

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(currentBalance))
	mock.ExpectExec(`UPDATE wallets SET balance = \$1 WHERE id = \$2`).
		WithArgs(newBalance, walletID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.UpdateBalance(context.Background(), walletID, amount)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateBalance_InsufficientFunds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	walletID := "550e8400-e29b-41d4-a716-446655440000"
	currentBalance := int64(100)
	amount := int64(-500)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1 FOR UPDATE`).
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(currentBalance))
	mock.ExpectRollback()

	err = repo.UpdateBalance(context.Background(), walletID, amount)
	if err != appErr.ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateBalance_BeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	err = repo.UpdateBalance(context.Background(), "test-wallet", 100)
	if err == nil {
		t.Error("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetBalance_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	walletID := "550e8400-e29b-41d4-a716-446655440000"
	expectedBalance := int64(5000)

	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1`).
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(expectedBalance))

	balance, err := repo.GetBalance(context.Background(), walletID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if balance != expectedBalance {
		t.Errorf("expected balance %d, got %d", expectedBalance, balance)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetBalance_WalletNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	walletID := "550e8400-e29b-41d4-a716-446655440000"

	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1`).
		WithArgs(walletID).
		WillReturnError(sql.ErrNoRows)

	balance, err := repo.GetBalance(context.Background(), walletID)
	if err != appErr.ErrWalletNotFound {
		t.Errorf("expected ErrWalletNotFound, got %v", err)
	}

	if balance != 0 {
		t.Errorf("expected balance 0, got %d", balance)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetBalance_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	walletID := "550e8400-e29b-41d4-a716-446655440000"

	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id = \$1`).
		WithArgs(walletID).
		WillReturnError(sql.ErrConnDone)

	_, err = repo.GetBalance(context.Background(), walletID)
	if err == nil {
		t.Error("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
