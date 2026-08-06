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
	query := `
		INSERT INTO portfolios (user_id, cash_balance)
		VALUES ($1, $2)
		RETURNING id, user_id, cash_balance, season_id, league_id, season_ref_id, created_at
	`
	var p model.Portfolio
	err := r.server.DB.Pool.QueryRow(ctx, query, userID, startingBalance).Scan(
		&p.ID, &p.UserID, &p.CashBalance, &p.SeasonID, &p.LeagueID, &p.SeasonRefID, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create portfolio: %w", err)
	}
	return &p, nil
}
func (r *PortfolioRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Portfolio, error) {
	query := `
		SELECT id, user_id, cash_balance, season_id, league_id, season_ref_id, created_at
		FROM portfolios
		WHERE user_id = $1 AND league_id IS NULL
	`
	var p model.Portfolio
	err := r.server.DB.Pool.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.CashBalance, &p.SeasonID, &p.LeagueID, &p.SeasonRefID, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get portfolio by user id: %w", err)
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

func (r *PortfolioRepository) ListAll(ctx context.Context) ([]model.Portfolio, error) {
	query := `
		SELECT id, user_id, cash_balance, season_id, league_id, season_ref_id, created_at
		FROM portfolios
		WHERE league_id IS NULL
		ORDER BY created_at ASC
	`
	rows, err := r.server.DB.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all portfolios: %w", err)
	}
	defer rows.Close()
	var portfolios []model.Portfolio
	for rows.Next() {
		var p model.Portfolio
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.CashBalance, &p.SeasonID, &p.LeagueID, &p.SeasonRefID, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan portfolio: %w", err)
		}
		portfolios = append(portfolios, p)
	}
	return portfolios, nil
}

func (r *PortfolioRepository) CreateLeaguePortfolio(
	ctx context.Context,
	tx pgx.Tx,
	userID, leagueID, seasonID uuid.UUID,
	startingBalance int,
) (*model.Portfolio, error) {
	query := `
		INSERT INTO portfolios (user_id, cash_balance, season_id, league_id, season_ref_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, cash_balance, season_id, league_id, season_ref_id, created_at
	`
	var p model.Portfolio
	err := tx.QueryRow(
		ctx, query,
		userID, startingBalance, seasonID.String(), leagueID, seasonID,
	).Scan(
		&p.ID, &p.UserID, &p.CashBalance, &p.SeasonID, &p.LeagueID, &p.SeasonRefID, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create lea	gue portfolio: %w", err)
	}
	return &p, nil
}

func (r *PortfolioRepository) DeleteLeaguePortfolios(ctx context.Context, tx pgx.Tx, userID, leagueID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM portfolios WHERE user_id = $1 AND league_id = $2`)
	if err != nil {
		return fmt.Errorf("delete league portfolios : %w", err)
	}
	return nil
}
