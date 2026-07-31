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

type LeagueRepository struct {
	server *server.Server
}

func NewLeagueRepository(s *server.Server) *LeagueRepository {
	return &LeagueRepository{
		server: s,
	}
}

func (r *LeagueRepository) Create(ctx context.Context, tx pgx.Tx, name string, createdBy uuid.UUID, inviteCode string, maxMembers int) (*model.League, error) {
	query := `INSERT INTO leagues (name, created_by, invite_code, max_members) VALUES ($1, $2, $3, $4) RETURNING id, name, created_by, invite_code, max_members, is_public, status, created_at`

	var l model.League
	err := tx.QueryRow(ctx, query, name, createdBy, inviteCode, maxMembers).Scan(
		&l.ID, &l.Name, &l.CreatedBy, &l.InviteCode, &l.MaxMembers, &l.IsPublic, &l.Status, &l.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create league : %w", err)
	}

	return &l, nil

}

func (r *LeagueRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.League, error) {
	query := `SELECT id, name, created_by, invite_code, max_members, is_public, status, created_at FROM leagues WHERE id = $1`

	var l model.League
	err := r.server.DB.Pool.QueryRow(ctx, query, id).Scan(&l.ID, &l.Name, &l.CreatedBy, &l.InviteCode, &l.MaxMembers, &l.IsPublic, &l.Status, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get league by id:%w", err)
	}
	return &l, nil
}

func (r *LeagueRepository) GetByInviteCode(ctx context.Context, code string) (*model.League, error) {
	query := `SELECT id, name, created_by, invite_code, max_members, is_public, status, created_at FROM leagues WHERE invite_code = $1`

	var l model.League
	err := r.server.DB.Pool.QueryRow(ctx, query, code).Scan(&l.ID, &l.Name, &l.CreatedBy, &l.InviteCode, &l.MaxMembers, &l.IsPublic, &l.Status, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get league by invite code : %w", err)
	}
	return &l, nil
}

func (r *LeagueRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.League, error) {
	query := `SELECT id, name, created_by, invite_code, max_members, is_public, status, created_at FROM leagues l INNER JOIN 
			league_members m ON l.id = m.league_id WHERE m.user_id = $1 ORDER BY l.created_at DESC`

	rows, err := r.server.DB.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list leagues by user : %w", err)
	}

	defer rows.Close()

	var leagues []model.League
	for rows.Next() {
		var l model.League
		if err := rows.Scan(&l.ID, &l.Name, &l.CreatedBy, &l.InviteCode, &l.MaxMembers, &l.IsPublic, &l.Status, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("sccan league : %w", err)
		}
		leagues = append(leagues, l)
	}
	return leagues, nil
}

func (r *LeagueRepository) AddMember(
	ctx context.Context,
	tx pgx.Tx,
	leagueID, userID uuid.UUID,
	role string,
) (*model.LeagueMember, error) {
	query := `
		INSERT INTO league_members (league_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, league_id, user_id, role, joined_at
	`
	var m model.LeagueMember
	err := tx.QueryRow(ctx, query, leagueID, userID, role).Scan(
		&m.ID, &m.LeagueID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add league member: %w", err)
	}
	return &m, nil
}
func (r *LeagueRepository) CountMembers(ctx context.Context, leagueID uuid.UUID) (int, error) {
	var count int
	err := r.server.DB.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM league_members WHERE league_id = $1`, leagueID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count league members: %w", err)
	}
	return count, nil
}
func (r *LeagueRepository) IsMember(ctx context.Context, leagueID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.server.DB.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM league_members WHERE league_id = $1 AND user_id = $2
		)
	`, leagueID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("is league member: %w", err)
	}
	return exists, nil
}
func (r *LeagueRepository) ListMembers(ctx context.Context, leagueID uuid.UUID) ([]model.LeagueMember, error) {
	query := `
		SELECT id, league_id, user_id, role, joined_at
		FROM league_members
		WHERE league_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := r.server.DB.Pool.Query(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("list league members: %w", err)
	}
	defer rows.Close()
	var members []model.LeagueMember
	for rows.Next() {
		var m model.LeagueMember
		if err := rows.Scan(&m.ID, &m.LeagueID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan league member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}
