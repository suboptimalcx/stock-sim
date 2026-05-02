package service

import (
	"context"
	"errors"
	"stock-sim/internal/model"
	"testing"
)

type MockRepository struct {
	TradeBuyFunc       func(ctx context.Context, walletID, stockName string) error
	TradeSellFunc      func(ctx context.Context, walletID, stockName string) error
	GetWalletFunc      func(ctx context.Context, walletID string) (model.Wallet, error)
	GetWalletStockFunc func(ctx context.Context, walletID, stockName string) (int, error)
	GetBankStateFunc   func(ctx context.Context) ([]model.Stock, error)
	SetBankStateFunc   func(ctx context.Context, stocks []model.Stock) error
	GetLogsFunc        func(ctx context.Context) ([]model.LogEntry, error)
}

func (m *MockRepository) TradeBuy(ctx context.Context, walletID, stockName string) error {
	if m.TradeBuyFunc != nil {
		return m.TradeBuyFunc(ctx, walletID, stockName)
	}
	return nil
}

func (m *MockRepository) TradeSell(ctx context.Context, walletID, stockName string) error {
	if m.TradeSellFunc != nil {
		return m.TradeSellFunc(ctx, walletID, stockName)
	}
	return nil
}

func (m *MockRepository) GetWallet(ctx context.Context, walletID string) (model.Wallet, error) {
	if m.GetWalletFunc != nil {
		return m.GetWalletFunc(ctx, walletID)
	}
	return model.Wallet{ID: walletID}, nil
}

func (m *MockRepository) GetWalletStock(ctx context.Context, walletID, stockName string) (int, error) {
	if m.GetWalletStockFunc != nil {
		return m.GetWalletStockFunc(ctx, walletID, stockName)
	}
	return 0, nil
}

func (m *MockRepository) GetBankState(ctx context.Context) ([]model.Stock, error) {
	if m.GetBankStateFunc != nil {
		return m.GetBankStateFunc(ctx)
	}
	return []model.Stock{}, nil
}

func (m *MockRepository) SetBankState(ctx context.Context, stocks []model.Stock) error {
	if m.SetBankStateFunc != nil {
		return m.SetBankStateFunc(ctx, stocks)
	}
	return nil
}

func (m *MockRepository) GetLogs(ctx context.Context) ([]model.LogEntry, error) {
	if m.GetLogsFunc != nil {
		return m.GetLogsFunc(ctx)
	}
	return []model.LogEntry{}, nil
}

func TestExecuteTrade_InvalidType(t *testing.T) {
	svc := NewMarketService(&MockRepository{})

	err := svc.ExecuteTrade(context.Background(), "user1", "AAPL", "not-a-type")

	if !errors.Is(err, model.ErrInvalidOperation) {
		t.Errorf("expected ErrInvalidOperation, got %v", err)
	}
}

func TestExecuteTrade_BuySuccess(t *testing.T) {
	mockRepo := &MockRepository{
		TradeBuyFunc: func(ctx context.Context, w, s string) error {
			return nil
		},
	}
	svc := NewMarketService(mockRepo)

	err := svc.ExecuteTrade(context.Background(), "user1", "AAPL", "buy")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestExecuteTrade_SellSuccess(t *testing.T) {
	mockRepo := &MockRepository{
		TradeSellFunc: func(ctx context.Context, w, s string) error {
			return nil
		},
	}
	svc := NewMarketService(mockRepo)

	err := svc.ExecuteTrade(context.Background(), "user1", "AAPL", "sell")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestExecuteTrade_BuyError(t *testing.T) {
	mockRepo := &MockRepository{
		TradeBuyFunc: func(ctx context.Context, w, s string) error {
			return model.ErrStockNotFound
		},
	}
	svc := NewMarketService(mockRepo)

	err := svc.ExecuteTrade(context.Background(), "user1", "AAPL", "buy")

	if !errors.Is(err, model.ErrStockNotFound) {
		t.Errorf("expected ErrStockNotFound, got %v", err)
	}
}

func TestExecuteTrade_SellError(t *testing.T) {
	mockRepo := &MockRepository{
		TradeSellFunc: func(ctx context.Context, w, s string) error {
			return model.ErrInsufficientStock
		},
	}
	svc := NewMarketService(mockRepo)

	err := svc.ExecuteTrade(context.Background(), "user1", "AAPL", "sell")

	if !errors.Is(err, model.ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestGetWallet(t *testing.T) {
	expectedWallet := model.Wallet{
		ID: "user1",
		Stocks: []model.Stock{
			{Name: "AAPL", Quantity: 10},
		},
	}
	mockRepo := &MockRepository{
		GetWalletFunc: func(ctx context.Context, walletID string) (model.Wallet, error) {
			return expectedWallet, nil
		},
	}
	svc := NewMarketService(mockRepo)

	wallet, err := svc.GetWallet(context.Background(), "user1")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if wallet.ID != expectedWallet.ID {
		t.Errorf("expected wallet ID %s, got %s", expectedWallet.ID, wallet.ID)
	}
	if len(wallet.Stocks) != len(expectedWallet.Stocks) {
		t.Errorf("expected %d stocks, got %d", len(expectedWallet.Stocks), len(wallet.Stocks))
	}
}

func TestGetWalletStock(t *testing.T) {
	mockRepo := &MockRepository{
		GetWalletStockFunc: func(ctx context.Context, walletID, stockName string) (int, error) {
			if stockName == "AAPL" {
				return 5, nil
			}
			return 0, model.ErrStockNotFound
		},
	}
	svc := NewMarketService(mockRepo)

	qty, err := svc.GetWalletStock(context.Background(), "user1", "AAPL")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if qty != 5 {
		t.Errorf("expected quantity 5, got %d", qty)
	}

	_, err = svc.GetWalletStock(context.Background(), "user1", "GOOG")
	if !errors.Is(err, model.ErrStockNotFound) {
		t.Errorf("expected ErrStockNotFound, got %v", err)
	}
}

func TestGetBankState(t *testing.T) {
	expectedStocks := []model.Stock{
		{Name: "AAPL", Quantity: 100},
		{Name: "GOOG", Quantity: 50},
	}
	mockRepo := &MockRepository{
		GetBankStateFunc: func(ctx context.Context) ([]model.Stock, error) {
			return expectedStocks, nil
		},
	}
	svc := NewMarketService(mockRepo)

	stocks, err := svc.GetBankState(context.Background())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(stocks) != len(expectedStocks) {
		t.Errorf("expected %d stocks, got %d", len(expectedStocks), len(stocks))
	}
}

func TestSetBankState(t *testing.T) {
	stocks := []model.Stock{
		{Name: "AAPL", Quantity: 100},
	}
	mockRepo := &MockRepository{
		SetBankStateFunc: func(ctx context.Context, s []model.Stock) error {
			if len(s) != 1 || s[0].Name != "AAPL" {
				return errors.New("unexpected stocks")
			}
			return nil
		},
	}
	svc := NewMarketService(mockRepo)

	err := svc.SetBankState(context.Background(), stocks)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGetLogs(t *testing.T) {
	expectedLogs := []model.LogEntry{
		{Type: "buy", WalletID: "user1", StockName: "AAPL"},
	}
	mockRepo := &MockRepository{
		GetLogsFunc: func(ctx context.Context) ([]model.LogEntry, error) {
			return expectedLogs, nil
		},
	}
	svc := NewMarketService(mockRepo)

	logs, err := svc.GetLogs(context.Background())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(logs) != len(expectedLogs) {
		t.Errorf("expected %d logs, got %d", len(expectedLogs), len(logs))
	}
}
