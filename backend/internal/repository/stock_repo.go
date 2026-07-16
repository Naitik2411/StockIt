package repository

import (
	"context"
	"fmt"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/server"
)

type StockRepository struct {
	server *server.Server
}

func NewStockRepository(s *server.Server) *StockRepository {
	return &StockRepository{server: s}
}

func (r *StockRepository) Upsert(ctx context.Context, stock *model.Stock) error {
	query := `
		INSERT INTO stocks (ticker, name, current_price, price_change_pct, last_synced_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (ticker) DO UPDATE SET
			name = EXCLUDED.name,
			current_price = EXCLUDED.current_price,
			price_change_pct = EXCLUDED.price_change_pct,
			last_synced_at = EXCLUDED.last_synced_at
	`
	_, err := r.server.DB.Pool.Exec(ctx, query, stock.Ticker,
		stock.Name,
		stock.CurrentPrice,
		stock.PriceChangePct,
		stock.LastSyncedAt)
	if err != nil {
		return fmt.Errorf("upsert stock : %w", err)
	}
	return nil
}

func (r *StockRepository) GetByTicker(ctx context.Context, ticker string) (*model.Stock, error) {
	query := `SELECT ticker, name, current_price, price_change_pct, last_synced_at FROM stocks WHERE ticker = $1`
	var stock model.Stock
	err := r.server.DB.Pool.QueryRow(ctx, query, ticker).Scan(
		&stock.Ticker, &stock.Name, &stock.CurrentPrice, &stock.PriceChangePct, &stock.LastSyncedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get stock by ticker: %w", err)
	}
	return &stock, nil

}
