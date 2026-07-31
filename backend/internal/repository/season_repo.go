package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SeasonRepository struct {
	server *server.Server
}

func NewSeasonRepository(s *server.Server) *SeasonRepository {
	return &SeasonRepository{
		server: s,
	}
}

func (r *SeasonRepository) Create(
	ctx context.Context,
	tx pgx.Tx,
	leagueID uuid.UUID,
	seasonNumber int,
	startDate, endDate time.Time,
) (*model.Season, error) {
	query := `
		INSERT INTO seasons (league_id, season_number, start_date, end_date, status)
		VALUES ($1, $2, $3, $4, 'active')
		RETURNING id, league_id, season_number, start_date, end_date, status, created_at
	`
	var s model.Season
	err := tx.QueryRow(ctx, query, leagueID, seasonNumber, startDate, endDate).Scan(
		&s.ID, &s.LeagueID, &s.SeasonNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create season: %w", err)
	}
	return &s, nil
}
func (r *SeasonRepository) GetCurrentActive(ctx context.Context, leagueID uuid.UUID) (*model.Season, error) {
	query := `
		SELECT id, league_id, season_number, start_date, end_date, status, created_at
		FROM seasons
		WHERE league_id = $1 AND status = 'active'
		ORDER BY season_number DESC
		LIMIT 1
	`
	var s model.Season
	err := r.server.DB.Pool.QueryRow(ctx, query, leagueID).Scan(
		&s.ID, &s.LeagueID, &s.SeasonNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get current season: %w", err)
	}
	return &s, nil
}
