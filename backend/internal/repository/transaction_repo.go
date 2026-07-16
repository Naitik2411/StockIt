package repository

import (
	"context"
	"fmt"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type TransactionRepository struct {
	server *server.Server
}

func NewTransactionRepository(s *server.Server) *TransactionRepository {
	return &TransactionRepository{
		server: s,
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx pgx.Tx, portfolioID uuid.UUID, ticker string, tradeType string, shares, price, totalValue decimal.Decimal) error {
	query := `INSERT INTO transactions (portfolio_id, ticker, type, shares, price_at_trade, total_value) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tx.Exec(ctx, query, portfolioID, ticker, tradeType, shares, price, totalValue)
	if err != nil {
		return fmt.Errorf("create transactions : %w", err)
	}
	return nil
}

func (r *TransactionRepository) ListByPortfolio(ctx context.Context, portfolioID uuid.UUID, page, limit int) ([]model.Transaction, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := `SELECT id, portfolio_id, ticker, type, shares, price_at_trade, total_value, created_at
			FROM transactions where portfolio_id = $1 order by created_at desc LIMIT $2 OFFSET $3`

	rows, err := r.server.DB.Pool.Query(ctx, query, portfolioID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list transactions : %w", err)
	}
	defer rows.Close()
	var txns []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(&t.ID, &t.PortfolioID, &t.Ticker, &t.Type, &t.Shares, &t.PriceAtTrade, &t.TotalValue, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		txns = append(txns, t)
	}
	return txns, nil
}
