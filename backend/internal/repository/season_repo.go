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

func (r *SeasonRepository) FindExpired(ctx context.Context) ([]model.Season, error) {
	rows, err := r.server.DB.Pool.Query(ctx, `SELECT id, league_id, season_number, start_date, end_date, status, created_at
		FROM seasons
		WHERE status = 'active' AND end_date <= now()
		ORDER BY end_date ASC`)
	if err != nil {
		return nil, fmt.Errorf("find expired seasons: %w", err)
	}

	defer rows.Close()
	var seasons []model.Season
	for rows.Next() {
		var s model.Season
		if err := rows.Scan(
			&s.ID, &s.LeagueID, &s.SeasonNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan expired season: %w", err)
		}
		seasons = append(seasons, s)
	}
	return seasons, nil
}

// GetForUpdate locks the season row for the duration of the transaction so two
// workers can never close the same season concurrently.
func (r *SeasonRepository) GetForUpdate(ctx context.Context, tx pgx.Tx, seasonID uuid.UUID) (*model.Season, error) {
	var s model.Season
	err := tx.QueryRow(ctx, `
		SELECT id, league_id, season_number, start_date, end_date, status, created_at
		FROM seasons WHERE id = $1
		FOR UPDATE
	`, seasonID).Scan(
		&s.ID, &s.LeagueID, &s.SeasonNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get season for update: %w", err)
	}
	return &s, nil
}

func (r *SeasonRepository) MarkCompleted(ctx context.Context, tx pgx.Tx, seasonID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE seasons SET status = 'completed' WHERE id = $1
	`, seasonID)
	if err != nil {
		return fmt.Errorf("mark season completed: %w", err)
	}
	return nil
}

func (r *SeasonRepository) MaxSeasonNumber(ctx context.Context, tx pgx.Tx, leagueID uuid.UUID) (int, error) {
	var maxNumber int
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(season_number), 0) FROM seasons WHERE league_id = $1
	`, leagueID).Scan(&maxNumber)
	if err != nil {
		return 0, fmt.Errorf("max season number: %w", err)
	}
	return maxNumber, nil
}

func (r *SeasonRepository) ListByLeague(ctx context.Context, leagueID uuid.UUID, page, limit int) ([]model.Season, int, error) {
	var total int

	if err := r.server.DB.Pool.QueryRow(ctx, `SELECT COUNT(*) from seasons WHERE league_id=$1`, leagueID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count seasons : %w", err)
	}
	offset := (page - 1) * limit

	rows, err := r.server.DB.Pool.Query(ctx, `SELECT id, league_id, season_number, start_date, end_date, status, created_at
				FROM seasons WHERE league_id = $1 ORDER BY season_number LIMIT $2 offset $3`, leagueID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list seasons by league:%w", err)
	}

	defer rows.Close()
	var seasons []model.Season
	for rows.Next() {
		var s model.Season
		if err := rows.Scan(&s.ID, &s.LeagueID, &s.SeasonNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan seasons : %w", err)
		}
		seasons = append(seasons, s)
	}
	return seasons, total, nil
}
