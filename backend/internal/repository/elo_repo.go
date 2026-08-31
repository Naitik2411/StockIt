package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ELORepository struct {
	server *server.Server
}

func NewELORepository(s *server.Server) *ELORepository {
	return &ELORepository{server: s}
}

type RatingUpdate struct {
	UserID uuid.UUID
	Rating int
}

// GetRatings returns current ratings for the given users. Users with no row are
// simply absent from the map — the caller applies the starting rating.
func (r *ELORepository) GetRatings(ctx context.Context, tx pgx.Tx, userIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	result := make(map[uuid.UUID]int, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT user_id, rating FROM elo_ratings WHERE user_id = ANY($1)
	`, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get elo ratings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var rating int
		if err := rows.Scan(&id, &rating); err != nil {
			return nil, fmt.Errorf("scan elo rating: %w", err)
		}
		result[id] = rating
	}
	return result, nil
}

// UpsertBulk writes new ratings, bumps seasons_played, and tracks peak_rating.
func (r *ELORepository) UpsertBulk(
	ctx context.Context,
	tx pgx.Tx,
	updates []RatingUpdate,
	startingRating int,
) error {
	if len(updates) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, u := range updates {
		batch.Queue(`
			INSERT INTO elo_ratings (user_id, rating, seasons_played, peak_rating, last_updated)
			VALUES ($1, $2, 1, GREATEST($2, $3), now())
			ON CONFLICT (user_id) DO UPDATE SET
				rating         = EXCLUDED.rating,
				seasons_played = elo_ratings.seasons_played + 1,
				peak_rating    = GREATEST(elo_ratings.peak_rating, EXCLUDED.rating),
				last_updated   = now()
		`, u.UserID, u.Rating, startingRating)
	}

	results := tx.SendBatch(ctx, batch)
	defer func() { _ = results.Close() }()

	for range updates {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("upsert elo rating: %w", err)
		}
	}
	return nil
}

func (r *ELORepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.ELORating, error) {
	var e model.ELORating
	err := r.server.DB.Pool.QueryRow(ctx, `
		SELECT user_id, rating, seasons_played, peak_rating, last_updated
		FROM elo_ratings WHERE user_id = $1
	`, userID).Scan(&e.UserID, &e.Rating, &e.SeasonsPlayed, &e.PeakRating, &e.LastUpdated)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get elo rating by user: %w", err)
	}
	return &e, nil
}

func (r *ELORepository) ListTop(ctx context.Context, page, limit int) ([]model.ELORating, int, error) {
	var total int
	if err := r.server.DB.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM elo_ratings`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count elo ratings: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := r.server.DB.Pool.Query(ctx, `
		SELECT user_id, rating, seasons_played, peak_rating, last_updated
		FROM elo_ratings
		ORDER BY rating DESC, seasons_played DESC, user_id ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list top elo ratings: %w", err)
	}
	defer rows.Close()

	var out []model.ELORating
	for rows.Next() {
		var e model.ELORating
		if err := rows.Scan(&e.UserID, &e.Rating, &e.SeasonsPlayed, &e.PeakRating, &e.LastUpdated); err != nil {
			return nil, 0, fmt.Errorf("scan elo rating: %w", err)
		}
		out = append(out, e)
	}
	return out, total, nil
}
