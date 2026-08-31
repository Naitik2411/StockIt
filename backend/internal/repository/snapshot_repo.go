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

type SnapshotRepository struct {
	server *server.Server
}

func NewSnapshotRepository(s *server.Server) *SnapshotRepository {
	return &SnapshotRepository{server: s}
}

type SnapshotInput struct {
	UserID     uuid.UUID
	FinalValue decimal.Decimal
	ReturnPct  decimal.Decimal
	Rank       int
}

// CreateBulk inserts one row per member. ON CONFLICT keeps the call idempotent if a season close is ever retried.
func (r *SnapshotRepository) CreateBulk(ctx context.Context, tx pgx.Tx, seasonID, leagueID uuid.UUID, rows []SnapshotInput) error {
	if len(rows) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, row := range rows {
		batch.Queue(
			`INSERT INTO season_snapshots (season_id, user_id, league_id, final_value, return_pct, rank)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (season_id, user_id) DO NOTHING
		`, seasonID, row.UserID, leagueID, row.FinalValue, row.ReturnPct, row.Rank)
	}

	results := tx.SendBatch(ctx, batch)
	defer func() { _ = results.Close() }()

	for range rows {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("insert season snapshot : %w", err)
		}
	}
	return nil
}

func (r *SnapshotRepository) ListBySeasonIDs(ctx context.Context, seasonIDs []uuid.UUID) (map[uuid.UUID][]model.SeasonSnapshot, error) {
	result := make(map[uuid.UUID][]model.SeasonSnapshot)
	if len(seasonIDs) == 0 {
		return result, nil
	}

	rows, err := r.server.DB.Pool.Query(ctx, `SELECT id, season_id, user_id, league_id, final_value, return_pct, rank, created_at
				FROM season_snapshots WHERE season_id=ANY($1) ORDER BY season_id, rank ASC`, seasonIDs)
	if err != nil {
		return nil, fmt.Errorf("list snapshots by season ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s model.SeasonSnapshot
		if err := rows.Scan(
			&s.ID, &s.SeasonID, &s.UserID, &s.LeagueID,
			&s.FinalValue, &s.ReturnPct, &s.Rank, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		result[s.SeasonID] = append(result[s.SeasonID], s)
	}
	return result, nil
}

func (r *SnapshotRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.SeasonSnapshot, error) {
	rows, err := r.server.DB.Pool.Query(ctx, `
		SELECT id, season_id, user_id, league_id, final_value, return_pct, rank, created_at
		FROM season_snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list snapshots by user: %w", err)
	}
	defer rows.Close()
	var out []model.SeasonSnapshot
	for rows.Next() {
		var s model.SeasonSnapshot
		if err := rows.Scan(
			&s.ID, &s.SeasonID, &s.UserID, &s.LeagueID,
			&s.FinalValue, &s.ReturnPct, &s.Rank, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		out = append(out, s)
	}
	return out, nil
}

