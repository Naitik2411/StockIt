package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type PositionRepository struct {
	server *server.Server
}

func NewPositionRepository(s *server.Server) *PositionRepository {
	return &PositionRepository{
		server: s,
	}
}

func (r *PositionRepository) GetByPortfolioAndTicker(ctx context.Context, tx pgx.Tx, portfolioID uuid.UUID, ticker string) (*model.Position, error) {
	query := `SELECT id, portfolio_id, ticker, shares, avg_buy_price, created_at, updated_at FROM positions where portfolio_id = $1 and ticker= $2`

	var p model.Position
	err := tx.QueryRow(ctx, query, portfolioID, ticker).Scan(&p.ID, &p.PortfolioID, &p.Ticker, &p.Shares, &p.AvgBuyPrice, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get positions : %w", err)
	}
	return &p, nil
}

func (r *PositionRepository) ListByPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]model.Position, error) {
	query := `SELECT id, portfolio_id, ticker, shares, avg_buy_price, created_at, updated_at from positions where portfolio_id=$1`

	rows, err := r.server.DB.Pool.Query(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("list positions : %w", err)
	}
	defer rows.Close()

	var positions []model.Position
	for rows.Next() {
		var p model.Position
		if err := rows.Scan(&p.ID, &p.PortfolioID, &p.Ticker, &p.Shares, &p.AvgBuyPrice, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan positions : %w", err)
		}
		positions = append(positions, p)
	}
	return positions, nil
}

func (r *PositionRepository) UpsertBuy(ctx context.Context, tx pgx.Tx, portfolioID uuid.UUID, ticker string, shares, price decimal.Decimal) error {

	query := `INSERT into positions (portfolio_id, ticker, shares, avg_buy_price) VALUES ($1, $2, $3, $4)
	ON CONFLICT (portfolio_id, ticker) DO UPDATE SET shares = position.shares+EXCLUDED.shares,
	avg_buy_price = (
				(positions.shares * positions.avg_buy_price) + (EXCLUDED.shares * EXCLUDED.avg_buy_price)
			) / (positions.shares + EXCLUDED.shares),
			updated_at = now()
	`
	_, err := tx.Exec(ctx, query, portfolioID, ticker, shares, price)
	if err != nil {
		return fmt.Errorf("upsert any positions : %w", err)
	}
	return nil
}

func (r *PositionRepository) UpdateShares(ctx context.Context, tx pgx.Tx, positionID uuid.UUID, newShares decimal.Decimal) error {
	_, err := tx.Exec(ctx, `
		UPDATE positions SET shares = $1, updated_at = now() WHERE id = $2
	`, newShares, positionID)
	if err != nil {
		return fmt.Errorf("update shares: %w", err)
	}
	return nil
}

func (r *PositionRepository) Delete(ctx context.Context, tx pgx.Tx, positionID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM positions WHERE id = $1`, positionID)
	if err != nil {
		return fmt.Errorf("delete position: %w", err)
	}
	return nil
}
