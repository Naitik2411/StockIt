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

type UserRepository struct {
	server *server.Server
}

func NewUserRepository(s *server.Server) *UserRepository {
	return &UserRepository{
		server: s,
	}
}

func (r *UserRepository) GetByClerkID(ctx context.Context, clerkUserID string) (*model.User, error) {
	query := `
		SELECT id, clerk_user_id, username, email, created_at, updated_at
		FROM users
		WHERE clerk_user_id = $1
	`
	var user model.User
	err := r.server.DB.Pool.QueryRow(ctx, query, clerkUserID).Scan(
		&user.ID,
		&user.ClerkUserID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found — not an error
		}
		return nil, fmt.Errorf("get user by clerk id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, clerkUserID string) (*model.User, error) {
	query := `INSERT INTO users (clerk_user_id) VALUES ($1) RETURNING id, clerk_user_id, username, email, created_at, updated_at`

	var user model.User
	err := r.server.DB.Pool.QueryRow(ctx, query, clerkUserID).Scan(
		&user.ID,
		&user.ClerkUserID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `SELECT id, clerk_user_id, username, email, created_at, updated_at
		FROM users
		WHERE id = $1`

	var user model.User

	err := r.server.DB.Pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.ClerkUserID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) ListByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*model.User, error) {
	result := make(map[uuid.UUID]*model.User)
	if len(ids) == 0 {
		return result, nil
	}
	query := `
		SELECT id, clerk_user_id, username, email, created_at, updated_at
		FROM users
		WHERE id = ANY($1)
	`
	rows, err := r.server.DB.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("list users by ids: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.ClerkUserID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		u := user
		result[user.ID] = &u
	}
	return result, nil
}
