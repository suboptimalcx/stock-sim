package repository

import (
	"context"
	"errors"
	"stock-sim/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) TradeBuy(ctx context.Context, walletID, stockName string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var bankQty int
	err = tx.QueryRow(ctx, "SELECT quantity FROM stocks WHERE name = $1 FOR UPDATE", stockName).Scan(&bankQty)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrStockNotFound
		}
		return err
	}
	if bankQty <= 0 {
		return model.ErrInsufficientStock
	}

	_, err = tx.Exec(ctx, "UPDATE stocks SET quantity = quantity - 1 WHERE name = $1", stockName)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO wallet_stocks (wallet_id, stock_name, quantity) 
        VALUES ($1, $2, 1) 
        ON CONFLICT (wallet_id, stock_name) 
        DO UPDATE SET quantity = wallet_stocks.quantity + 1`,
		walletID, stockName)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO audit_logs (type, wallet_id, stock_name) VALUES ($1, $2, $3)", "buy", walletID, stockName)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) TradeSell(ctx context.Context, walletID, stockName string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lock int
	err = tx.QueryRow(ctx, "SELECT 1 FROM stocks WHERE name = $1 FOR UPDATE", stockName).Scan(&lock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrStockNotFound
		}
		return err
	}

	var walletQty int
	err = tx.QueryRow(ctx, "SELECT quantity FROM wallet_stocks WHERE wallet_id = $1 AND stock_name = $2 FOR UPDATE", walletID, stockName).Scan(&walletQty)
	if err != nil || walletQty <= 0 {
		return model.ErrInsufficientStock
	}

	_, err = tx.Exec(ctx, "UPDATE wallet_stocks SET quantity = quantity - 1 WHERE wallet_id = $1 AND stock_name = $2", walletID, stockName)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE stocks SET quantity = quantity + 1 WHERE name = $1", stockName)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO audit_logs (type, wallet_id, stock_name) VALUES ($1, $2, $3)", "sell", walletID, stockName)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetWallet(ctx context.Context, walletID string) (model.Wallet, error) {
	wallet := model.Wallet{
		ID:     walletID,
		Stocks: make([]model.Stock, 0),
	}

	rows, err := r.db.Query(ctx, "SELECT stock_name, quantity FROM wallet_stocks WHERE wallet_id = $1", walletID)
	if err != nil {
		return wallet, err
	}
	defer rows.Close()

	for rows.Next() {
		var s model.Stock
		if err := rows.Scan(&s.Name, &s.Quantity); err != nil {
			return wallet, err
		}
		wallet.Stocks = append(wallet.Stocks, s)
	}

	return wallet, rows.Err()
}

func (r *PostgresRepository) GetWalletStock(ctx context.Context, walletID, stockName string) (int, error) {
	var qty int
	err := r.db.QueryRow(ctx, "SELECT quantity FROM wallet_stocks WHERE wallet_id = $1 AND stock_name = $2", walletID, stockName).Scan(&qty)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, model.ErrStockNotFound
		}
		return 0, err
	}
	return qty, nil
}

func (r *PostgresRepository) GetBankState(ctx context.Context) ([]model.Stock, error) {
	stocks := make([]model.Stock, 0)
	rows, err := r.db.Query(ctx, "SELECT name, quantity FROM stocks")
	if err != nil {
		return stocks, err
	}
	defer rows.Close()

	for rows.Next() {
		var s model.Stock
		if err := rows.Scan(&s.Name, &s.Quantity); err != nil {
			return stocks, err
		}
		stocks = append(stocks, s)
	}

	return stocks, rows.Err()
}

func (r *PostgresRepository) SetBankState(ctx context.Context, stocks []model.Stock) error {
	if len(stocks) == 0 {
		return nil
	}

	names := make([]string, len(stocks))
	quantities := make([]int, len(stocks))
	for i, s := range stocks {
		names[i] = s.Name
		quantities[i] = s.Quantity
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        INSERT INTO stocks (name, quantity)
        SELECT unnest($1::text[]), unnest($2::int[])
        ON CONFLICT (name)
        DO UPDATE SET quantity = EXCLUDED.quantity`,
		names, quantities)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetLogs(ctx context.Context) ([]model.LogEntry, error) {
	logs := make([]model.LogEntry, 0)
	rows, err := r.db.Query(ctx, "SELECT type, wallet_id, stock_name FROM audit_logs ORDER BY created_at ASC LIMIT 10000")
	if err != nil {
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		var l model.LogEntry
		if err := rows.Scan(&l.Type, &l.WalletID, &l.StockName); err != nil {
			return logs, err
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}
