package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"stock-sim/internal/model"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MarketService interface {
	ExecuteTrade(ctx context.Context, walletID, stockName, tradeType string) error
	GetWallet(ctx context.Context, walletID string) (model.Wallet, error)
	GetWalletStock(ctx context.Context, walletID, stockName string) (int, error)
	GetBankState(ctx context.Context) ([]model.Stock, error)
	SetBankState(ctx context.Context, stocks []model.Stock) error
	GetLogs(ctx context.Context) ([]model.LogEntry, error)
}

type Handler struct {
	svc MarketService
}

func NewHandler(svc MarketService) *chi.Mux {
	h := &Handler{svc: svc}
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/wallets/{wallet_id}/stocks/{stock_name}", h.Trade)
	r.Get("/wallets/{wallet_id}", h.GetWallet)
	r.Get("/wallets/{wallet_id}/stocks/{stock_name}", h.GetWalletStock)
	r.Get("/stocks", h.GetStocks)
	r.Post("/stocks", h.SetStocks)
	r.Get("/log", h.GetLog)
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
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, model.ErrInsufficientStock):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "wallet_id")
	wallet, err := h.svc.GetWallet(r.Context(), walletID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

func (h *Handler) GetWalletStock(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "wallet_id")
	stockName := chi.URLParam(r, "stock_name")

	qty, err := h.svc.GetWalletStock(r.Context(), walletID, stockName)
	if err != nil {
		if errors.Is(err, model.ErrStockNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%d\n", qty)
}

func (h *Handler) GetStocks(w http.ResponseWriter, r *http.Request) {
	stocks, err := h.svc.GetBankState(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.BankState{Stocks: stocks})
}

func (h *Handler) SetStocks(w http.ResponseWriter, r *http.Request) {
	var req model.BankState
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.SetBankState(r.Context(), req.Stocks); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetLog(w http.ResponseWriter, r *http.Request) {
	logs, err := h.svc.GetLogs(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.LogResponse{Log: logs})
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
