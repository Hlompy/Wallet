package handler

import (
	"context"
	"encoding/json"
	"net/http"

	appErr "github.com/Hlompy/Wallet/internal/errors"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type WalletService interface {
	Process(ctx context.Context, walletID, op string, amount int64) error
	Balance(ctx context.Context, walletID string) (int64, error)
}

type Handler struct {
	service WalletService // Изменено на интерфейс
}

func New(service WalletService) *Handler { // Изменено на интерфейс
	return &Handler{service: service}
}

type walletRequest struct {
	WalletID string `json:"walletId"`
	OpType   string `json:"operationType"`
	Amount   int64  `json:"amount"`
}

type walletResponse struct {
	WalletID string `json:"walletId"`
	Balance  int64  `json:"balance"`
}

func (h *Handler) PostWallet(w http.ResponseWriter, r *http.Request) {
	var req walletRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(req.WalletID); err != nil {
		http.Error(w, "invalid walletId", http.StatusBadRequest)
		return
	}

	err := h.service.Process(
		r.Context(),
		req.WalletID,
		req.OpType,
		req.Amount,
	)

	if err != nil {
		switch err {
		case appErr.ErrInsufficientFunds:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case appErr.ErrInvalidOperation:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case appErr.ErrWalletNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	balance, err := h.service.Balance(r.Context(), req.WalletID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := walletResponse{
		WalletID: req.WalletID,
		Balance:  balance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid walletId", http.StatusBadRequest)
		return
	}

	balance, err := h.service.Balance(r.Context(), id)
	if err != nil {
		switch err {
		case appErr.ErrWalletNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(map[string]int64{
		"balance": balance,
	})
}
