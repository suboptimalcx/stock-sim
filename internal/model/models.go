package model

import "errors"

var (
	ErrStockNotFound     = errors.New("stock not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidOperation  = errors.New("invalid operation")
)

type TradeRequest struct {
	Type string `json:"type"` // buy or sell
}

type Stock struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type BankState struct {
	Stocks []Stock `json:"stocks"`
}

type Wallet struct {
	ID     string  `json:"id"`
	Stocks []Stock `json:"stocks"`
}

type LogEntry struct {
	Type      string `json:"type"`
	WalletID  string `json:"wallet_id"`
	StockName string `json:"stock_name"`
}

type LogResponse struct {
	Log []LogEntry `json:"log"`
}
