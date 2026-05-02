package service

import (
	"context"
	"stock-sim/internal/model"
)

type Repository interface {
	Trade(ctx context.Context, walletID, stockName, tradeType string) error
	GetWallet(ctx context.Context, walletID string) (model.Wallet, error)
	GetWalletStock(ctx context.Context, walletID, stockName string) (int, error)
	GetBankState(ctx context.Context) ([]model.Stock, error)
	SetBankState(ctx context.Context, stocks []model.Stock) error
	GetLogs(ctx context.Context) ([]model.LogEntry, error)
}

type MarketService struct {
	repo Repository
}

func NewMarketService(repo Repository) *MarketService {
	return &MarketService{repo: repo}
}

func (s *MarketService) ExecuteTrade(ctx context.Context, walletID, stockName, tradeType string) error {
	if tradeType != "buy" && tradeType != "sell" {
		return model.ErrInvalidOperation
	}
	return s.repo.Trade(ctx, walletID, stockName, tradeType)
}

func (s *MarketService) GetWallet(ctx context.Context, walletID string) (model.Wallet, error) {
	return s.repo.GetWallet(ctx, walletID)
}

func (s *MarketService) GetWalletStock(ctx context.Context, walletID, stockName string) (int, error) {
	return s.repo.GetWalletStock(ctx, walletID, stockName)
}

func (s *MarketService) GetBankState(ctx context.Context) ([]model.Stock, error) {
	return s.repo.GetBankState(ctx)
}

func (s *MarketService) SetBankState(ctx context.Context, stocks []model.Stock) error {
	return s.repo.SetBankState(ctx, stocks)
}

func (s *MarketService) GetLogs(ctx context.Context) ([]model.LogEntry, error) {
	return s.repo.GetLogs(ctx)
}
