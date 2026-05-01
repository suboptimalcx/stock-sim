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

func (r *PostgresRepository) Trade(ctx context.Context, walletID, stockName, tradeType string) error {
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

	switch tradeType {
	case "buy":
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

	case "sell":
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
	default:
		return model.ErrInvalidOperation
	}

	_, err = tx.Exec(ctx, "INSERT INTO audit_logs (type, wallet_id, stock_name) VALUES ($1, $2, $3)", tradeType, walletID, stockName)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
