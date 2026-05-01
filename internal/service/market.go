package service

import (
	"context"
	"stock-sim/internal/model"
)

type Repository interface {
	Trade(ctx context.Context, walletID, stockName, tradeType string) error
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
