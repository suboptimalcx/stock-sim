package repository

import (
	"context"
	"os"
	"stock-sim/internal/model"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/mydb?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("Skipping integration tests: database unavailable (%v)", err)
	}

	_, err = pool.Exec(context.Background(), `
		DROP TABLE IF EXISTS audit_logs, wallet_stocks, stocks CASCADE;
		CREATE TABLE stocks (
			name VARCHAR(255) PRIMARY KEY,
			quantity INT NOT NULL CHECK (quantity >= 0)
		);
		CREATE TABLE wallet_stocks (
			wallet_id VARCHAR(255) NOT NULL,
			stock_name VARCHAR(255) REFERENCES stocks(name),
			quantity INT NOT NULL CHECK (quantity >= 0),
			PRIMARY KEY (wallet_id, stock_name)
		);
		CREATE TABLE audit_logs (
			id SERIAL PRIMARY KEY,
			type VARCHAR(10) NOT NULL,
			wallet_id VARCHAR(255) NOT NULL,
			stock_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Skipf("Skipping integration tests: failed to setup tables (%v)", err)
	}

	t.Cleanup(func() {
		pool.Exec(context.Background(), "DROP TABLE IF EXISTS audit_logs, wallet_stocks, stocks CASCADE")
		pool.Close()
	})

	return pool
}

func TestPostgresRepository_GetWallet(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresRepository(pool)

	ctx := context.Background()

	_, err := pool.Exec(ctx, "INSERT INTO stocks (name, quantity) VALUES ($1, $2)", "AAPL", 100)
	if err != nil {
		t.Fatalf("Failed to insert stock: %v", err)
	}
	_, err = pool.Exec(ctx, "INSERT INTO wallet_stocks (wallet_id, stock_name, quantity) VALUES ($1, $2, $3)", "user1", "AAPL", 10)
	if err != nil {
		t.Fatalf("Failed to insert wallet stock: %v", err)
	}

	wallet, err := repo.GetWallet(ctx, "user1")
	if err != nil {
		t.Errorf("GetWallet failed: %v", err)
	}

	if wallet.ID != "user1" {
		t.Errorf("expected wallet ID user1, got %s", wallet.ID)
	}

	if len(wallet.Stocks) != 1 {
		t.Errorf("expected 1 stock, got %d", len(wallet.Stocks))
	}

	if wallet.Stocks[0].Name != "AAPL" || wallet.Stocks[0].Quantity != 10 {
		t.Errorf("expected stock AAPL:10, got %s:%d", wallet.Stocks[0].Name, wallet.Stocks[0].Quantity)
	}
}

func TestPostgresRepository_GetWalletStock(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresRepository(pool)

	ctx := context.Background()

	_, err := pool.Exec(ctx, "INSERT INTO stocks (name, quantity) VALUES ($1, $2)", "AAPL", 100)
	if err != nil {
		t.Fatalf("Failed to insert stock: %v", err)
	}
	_, err = pool.Exec(ctx, "INSERT INTO wallet_stocks (wallet_id, stock_name, quantity) VALUES ($1, $2, $3)", "user1", "AAPL", 5)
	if err != nil {
		t.Fatalf("Failed to insert wallet stock: %v", err)
	}

	qty, err := repo.GetWalletStock(ctx, "user1", "AAPL")
	if err != nil {
		t.Errorf("GetWalletStock failed: %v", err)
	}

	if qty != 5 {
		t.Errorf("expected quantity 5, got %d", qty)
	}

	_, err = repo.GetWalletStock(ctx, "user1", "GOOG")
	if err == nil || err != model.ErrStockNotFound {
		t.Errorf("expected ErrStockNotFound, got %v", err)
	}
}

func TestPostgresRepository_GetBankState(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresRepository(pool)

	ctx := context.Background()

	_, err := pool.Exec(ctx, "INSERT INTO stocks (name, quantity) VALUES ($1, $2), ($3, $4)", "AAPL", 100, "GOOG", 50)
	if err != nil {
		t.Fatalf("Failed to insert stocks: %v", err)
	}

	stocks, err := repo.GetBankState(ctx)
	if err != nil {
		t.Errorf("GetBankState failed: %v", err)
	}

	if len(stocks) != 2 {
		t.Errorf("expected 2 stocks, got %d", len(stocks))
	}
}

func TestPostgresRepository_SetBankState(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresRepository(pool)

	ctx := context.Background()

	stocks := []model.Stock{
		{Name: "AAPL", Quantity: 100},
		{Name: "GOOG", Quantity: 50},
	}

	err := repo.SetBankState(ctx, stocks)
	if err != nil {
		t.Errorf("SetBankState failed: %v", err)
	}

	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM stocks").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count stocks: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 stocks in DB, got %d", count)
	}
}

func TestPostgresRepository_TradeBuy(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresRepository(pool)

	ctx := context.Background()

	_, err := pool.Exec(ctx, "INSERT INTO stocks (name, quantity) VALUES ($1, $2)", "AAPL", 10)
	if err != nil {
		t.Fatalf("Failed to insert stock: %v", err)
	}

	err = repo.TradeBuy(ctx, "user1", "AAPL")
	if err != nil {
		t.Errorf("TradeBuy failed: %v", err)
	}

	var qty int
	err = pool.QueryRow(ctx, "SELECT quantity FROM stocks WHERE name = $1", "AAPL").Scan(&qty)
	if err != nil {
		t.Errorf("Failed to get stock quantity: %v", err)
	}
	if qty != 9 {
		t.Errorf("expected stock quantity 9, got %d", qty)
	}

	err = pool.QueryRow(ctx, "SELECT quantity FROM wallet_stocks WHERE wallet_id = $1 AND stock_name = $2", "user1", "AAPL").Scan(&qty)
	if err != nil {
		t.Errorf("Failed to get wallet stock: %v", err)
	}
	if qty != 1 {
		t.Errorf("expected wallet stock 1, got %d", qty)
	}
}

func TestPostgresRepository_TradeSell(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresRepository(pool)

	ctx := context.Background()

	_, err := pool.Exec(ctx, "INSERT INTO stocks (name, quantity) VALUES ($1, $2)", "AAPL", 10)
	if err != nil {
		t.Fatalf("Failed to insert stock: %v", err)
	}
	_, err = pool.Exec(ctx, "INSERT INTO wallet_stocks (wallet_id, stock_name, quantity) VALUES ($1, $2, $3)", "user1", "AAPL", 5)
	if err != nil {
		t.Fatalf("Failed to insert wallet stock: %v", err)
	}

	err = repo.TradeSell(ctx, "user1", "AAPL")
	if err != nil {
		t.Errorf("TradeSell failed: %v", err)
	}

	var qty int
	err = pool.QueryRow(ctx, "SELECT quantity FROM stocks WHERE name = $1", "AAPL").Scan(&qty)
	if err != nil {
		t.Errorf("Failed to get stock quantity: %v", err)
	}
	if qty != 11 {
		t.Errorf("expected stock quantity 11, got %d", qty)
	}

	err = pool.QueryRow(ctx, "SELECT quantity FROM wallet_stocks WHERE wallet_id = $1 AND stock_name = $2", "user1", "AAPL").Scan(&qty)
	if err != nil {
		t.Errorf("Failed to get wallet stock: %v", err)
	}
	if qty != 4 {
		t.Errorf("expected wallet stock 4, got %d", qty)
	}
}

func TestPostgresRepository_GetLogs(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresRepository(pool)

	ctx := context.Background()

	_, err := pool.Exec(ctx, "INSERT INTO audit_logs (type, wallet_id, stock_name) VALUES ($1, $2, $3)", "buy", "user1", "AAPL")
	if err != nil {
		t.Fatalf("Failed to insert log: %v", err)
	}

	logs, err := repo.GetLogs(ctx)
	if err != nil {
		t.Errorf("GetLogs failed: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}

	if logs[0].Type != "buy" || logs[0].WalletID != "user1" || logs[0].StockName != "AAPL" {
		t.Errorf("log mismatch: %+v", logs[0])
	}
}

func TestPostgresRepository_ConcurrentBuy(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	repo.SetBankState(ctx, []model.Stock{{Name: "LIMIT", Quantity: 1}})

	const goroutines = 20
	errChan := make(chan error, goroutines)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			errChan <- repo.TradeBuy(ctx, "user-concurrent", "LIMIT")
		}(i)
	}

	wg.Wait()
	close(errChan)

	successCount := 0
	for err := range errChan {
		if err == nil {
			successCount++
		}
	}

	if successCount != 1 {
		t.Errorf("Race condition detected! Expected 1 success, got %d", successCount)
	}
}
