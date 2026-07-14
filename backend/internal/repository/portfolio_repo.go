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

type PortfolioRepository struct {
	server *server.Server
}

func NewPortfolioRepository(s *server.Server) *PortfolioRepository {
	return &PortfolioRepository{
		server: s,
	}
}

func (r *PortfolioRepository) Create(ctx context.Context, userID uuid.UUID, startingBalance int) (*model.Portfolio, error) {
	query := `INSERT INTO portfolios (user_id, cash_balance) VALUES ($1, $2) RETURNING id, user_id, cash_balance, season_id, created_at`
	var p model.Portfolio
	err := r.server.DB.Pool.QueryRow(ctx, query, userID, startingBalance).Scan(
		&p.ID, &p.UserID, &p.CashBalance, &p.SeasonID, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create portfolio : %w", err)
	}
	return &p, nil
}

func (r *PortfolioRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Portfolio, error) {
	query := `SELECT id, user_id, cash_balance, season_id, created_at FROM portfolios WHERE user_id = $1`

	var p model.Portfolio
	err := r.server.DB.Pool.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.CashBalance, &p.SeasonID, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get portfolio by user id : %w", err)
	}
	return &p, nil
}

func (r *PortfolioRepository) UpdateCash(ctx context.Context, tx pgx.Tx, portfolioID uuid.UUID, newCash decimal.Decimal) error {
	_, err := tx.Exec(ctx, `
		UPDATE portfolios SET cash_balance = $1 WHERE id = $2
	`, newCash, portfolioID)
	if err != nil {
		return fmt.Errorf("update cash: %w", err)
	}
	return nil
}
