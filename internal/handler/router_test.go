package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"stock-sim/internal/model"
	"strings"
	"testing"
)

type mockSvc struct {
	ExecuteTradeFunc   func(ctx context.Context, walletID, stockName, tradeType string) error
	GetWalletFunc      func(ctx context.Context, walletID string) (model.Wallet, error)
	GetWalletStockFunc func(ctx context.Context, walletID, stockName string) (int, error)
	GetBankStateFunc   func(ctx context.Context) ([]model.Stock, error)
	SetBankStateFunc   func(ctx context.Context, stocks []model.Stock) error
	GetLogsFunc        func(ctx context.Context) ([]model.LogEntry, error)
}

func (m *mockSvc) ExecuteTrade(ctx context.Context, walletID, stockName, tradeType string) error {
	if m.ExecuteTradeFunc != nil {
		return m.ExecuteTradeFunc(ctx, walletID, stockName, tradeType)
	}
	return nil
}

func (m *mockSvc) GetWallet(ctx context.Context, walletID string) (model.Wallet, error) {
	if m.GetWalletFunc != nil {
		return m.GetWalletFunc(ctx, walletID)
	}
	return model.Wallet{ID: walletID}, nil
}

func (m *mockSvc) GetWalletStock(ctx context.Context, walletID, stockName string) (int, error) {
	if m.GetWalletStockFunc != nil {
		return m.GetWalletStockFunc(ctx, walletID, stockName)
	}
	return 0, nil
}

func (m *mockSvc) GetBankState(ctx context.Context) ([]model.Stock, error) {
	if m.GetBankStateFunc != nil {
		return m.GetBankStateFunc(ctx)
	}
	return []model.Stock{}, nil
}

func (m *mockSvc) SetBankState(ctx context.Context, stocks []model.Stock) error {
	if m.SetBankStateFunc != nil {
		return m.SetBankStateFunc(ctx, stocks)
	}
	return nil
}

func (m *mockSvc) GetLogs(ctx context.Context) ([]model.LogEntry, error) {
	if m.GetLogsFunc != nil {
		return m.GetLogsFunc(ctx)
	}
	return []model.LogEntry{}, nil
}

func TestTradeHandler_Success(t *testing.T) {
	svc := &mockSvc{
		ExecuteTradeFunc: func(ctx context.Context, w, s, tt string) error {
			return nil
		},
	}
	router := NewHandler(svc)

	body := `{"type": "buy"}`
	req, _ := http.NewRequest("POST", "/wallets/user1/stocks/AAPL", strings.NewReader(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestTradeHandler_InvalidJSON(t *testing.T) {
	svc := &mockSvc{}
	router := NewHandler(svc)

	body := `invalid json`
	req, _ := http.NewRequest("POST", "/wallets/user1/stocks/AAPL", strings.NewReader(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestTradeHandler_StockNotFound(t *testing.T) {
	svc := &mockSvc{
		ExecuteTradeFunc: func(ctx context.Context, w, s, tt string) error {
			return model.ErrStockNotFound
		},
	}
	router := NewHandler(svc)

	body := `{"type": "buy"}`
	req, _ := http.NewRequest("POST", "/wallets/user1/stocks/AAPL", strings.NewReader(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestTradeHandler_InsufficientStock(t *testing.T) {
	svc := &mockSvc{
		ExecuteTradeFunc: func(ctx context.Context, w, s, tt string) error {
			return model.ErrInsufficientStock
		},
	}
	router := NewHandler(svc)

	body := `{"type": "buy"}`
	req, _ := http.NewRequest("POST", "/wallets/user1/stocks/AAPL", strings.NewReader(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestTradeHandler_InternalError(t *testing.T) {
	svc := &mockSvc{
		ExecuteTradeFunc: func(ctx context.Context, w, s, tt string) error {
			return model.ErrInvalidOperation
		},
	}
	router := NewHandler(svc)

	body := `{"type": "buy"}`
	req, _ := http.NewRequest("POST", "/wallets/user1/stocks/AAPL", strings.NewReader(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestGetWalletHandler(t *testing.T) {
	expectedWallet := model.Wallet{
		ID: "user1",
		Stocks: []model.Stock{
			{Name: "AAPL", Quantity: 10},
		},
	}
	svc := &mockSvc{
		GetWalletFunc: func(ctx context.Context, walletID string) (model.Wallet, error) {
			return expectedWallet, nil
		},
	}
	router := NewHandler(svc)

	req, _ := http.NewRequest("GET", "/wallets/user1", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var wallet model.Wallet
	if err := json.NewDecoder(rr.Body).Decode(&wallet); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if wallet.ID != expectedWallet.ID {
		t.Errorf("expected wallet ID %s, got %s", expectedWallet.ID, wallet.ID)
	}
}

func TestGetWalletStockHandler_Success(t *testing.T) {
	svc := &mockSvc{
		GetWalletStockFunc: func(ctx context.Context, w, s string) (int, error) {
			return 5, nil
		},
	}
	router := NewHandler(svc)

	req, _ := http.NewRequest("GET", "/wallets/user1/stocks/AAPL", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "5\n"
	if rr.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rr.Body.String())
	}
}

func TestGetWalletStockHandler_NotFound(t *testing.T) {
	svc := &mockSvc{
		GetWalletStockFunc: func(ctx context.Context, w, s string) (int, error) {
			return 0, model.ErrStockNotFound
		},
	}
	router := NewHandler(svc)

	req, _ := http.NewRequest("GET", "/wallets/user1/stocks/AAPL", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestGetStocksHandler(t *testing.T) {
	expectedStocks := []model.Stock{
		{Name: "AAPL", Quantity: 100},
	}
	svc := &mockSvc{
		GetBankStateFunc: func(ctx context.Context) ([]model.Stock, error) {
			return expectedStocks, nil
		},
	}
	router := NewHandler(svc)

	req, _ := http.NewRequest("GET", "/stocks", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var bankState model.BankState
	if err := json.NewDecoder(rr.Body).Decode(&bankState); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if len(bankState.Stocks) != len(expectedStocks) {
		t.Errorf("expected %d stocks, got %d", len(expectedStocks), len(bankState.Stocks))
	}
}

func TestSetStocksHandler_Success(t *testing.T) {
	svc := &mockSvc{
		SetBankStateFunc: func(ctx context.Context, stocks []model.Stock) error {
			return nil
		},
	}
	router := NewHandler(svc)

	body := `{"stocks": [{"name": "AAPL", "quantity": 100}]}`
	req, _ := http.NewRequest("POST", "/stocks", strings.NewReader(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestSetStocksHandler_InvalidJSON(t *testing.T) {
	svc := &mockSvc{}
	router := NewHandler(svc)

	body := `invalid json`
	req, _ := http.NewRequest("POST", "/stocks", strings.NewReader(body))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestGetLogHandler(t *testing.T) {
	expectedLogs := []model.LogEntry{
		{Type: "buy", WalletID: "user1", StockName: "AAPL"},
	}
	svc := &mockSvc{
		GetLogsFunc: func(ctx context.Context) ([]model.LogEntry, error) {
			return expectedLogs, nil
		},
	}
	router := NewHandler(svc)

	req, _ := http.NewRequest("GET", "/log", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var logResponse model.LogResponse
	if err := json.NewDecoder(rr.Body).Decode(&logResponse); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if len(logResponse.Log) != len(expectedLogs) {
		t.Errorf("expected %d logs, got %d", len(expectedLogs), len(logResponse.Log))
	}
}
