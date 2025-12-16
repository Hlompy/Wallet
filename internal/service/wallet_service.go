package service

import (
	"context"

	appErr "github.com/Hlompy/Wallet/internal/errors"
)

type WalletRepository interface {
	UpdateBalance(ctx context.Context, walletID string, amount int64) error
	GetBalance(ctx context.Context, walletID string) (int64, error)
}

type WalletService struct {
	repo WalletRepository
}

func New(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) Process(
	ctx context.Context,
	walletID string,
	op string,
	amount int64,
) error {

	if amount <= 0 {
		return appErr.ErrInvalidOperation
	}

	switch op {
	case "DEPOSIT":
		return s.repo.UpdateBalance(ctx, walletID, amount)
	case "WITHDRAW":
		return s.repo.UpdateBalance(ctx, walletID, -amount)
	default:
		return appErr.ErrInvalidOperation
	}
}

func (s *WalletService) Balance(ctx context.Context, walletID string) (int64, error) {
	return s.repo.GetBalance(ctx, walletID)
}
