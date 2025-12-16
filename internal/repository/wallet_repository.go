package repository

import (
	"context"
	"database/sql"

	appErr "github.com/Hlompy/Wallet/internal/errors"
)

type WalletRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) UpdateBalance(
	ctx context.Context,
	walletID string,
	amount int64,
) error {

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance int64
	err = tx.QueryRowContext(
		ctx,
		`SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`,
		walletID,
	).Scan(&balance)

	if err != nil {
		if err == sql.ErrNoRows {
			// 👉 СОЗДАЁМ КОШЕЛЁК
			if amount < 0 {
				return appErr.ErrWalletNotFound
			}

			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO wallets (id, balance) VALUES ($1, $2)`,
				walletID,
				amount,
			)
			if err != nil {
				return err
			}

			return tx.Commit()
		}
		return err
	}

	newBalance := balance + amount
	if newBalance < 0 {
		return appErr.ErrInsufficientFunds
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE wallets SET balance = $1 WHERE id = $2`,
		newBalance,
		walletID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *WalletRepository) GetBalance(
	ctx context.Context,
	walletID string,
) (int64, error) {

	var balance int64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT balance FROM wallets WHERE id = $1`,
		walletID,
	).Scan(&balance)

	if err == sql.ErrNoRows {
		return 0, appErr.ErrWalletNotFound
	}

	return balance, err
}
