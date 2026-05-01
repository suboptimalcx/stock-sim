package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"stock-sim/internal/model"
	"time"

	"github.com/go-chi/chi/v5"
)

type MarketService interface {
	ExecuteTrade(ctx context.Context, walletID, stockName, tradeType string) error
}

type Handler struct {
	svc MarketService
}

func NewHandler(svc MarketService) *chi.Mux {
	h := &Handler{svc: svc}
	r := chi.NewRouter()

	r.Post("/wallets/{wallet_id}/stocks/{stock_name}", h.Trade)

	r.Post("/chaos", h.Chaos)

	return r
}

func (h *Handler) Trade(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "wallet_id")
	stockName := chi.URLParam(r, "stock_name")

	var req model.TradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.svc.ExecuteTrade(r.Context(), walletID, stockName, req.Type)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrStockNotFound):
			http.Error(w, err.Error(), http.StatusNotFound) // 404
		case errors.Is(err, model.ErrInsufficientStock):
			http.Error(w, err.Error(), http.StatusBadRequest) // 400
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK) // 200 OK
}

func (h *Handler) Chaos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Instance is terminating...",
	})
	go func() {
		time.Sleep(100 * time.Millisecond)
		log.Println("CHAOS!!!!!!!!!")
		os.Exit(1)
	}()
}
