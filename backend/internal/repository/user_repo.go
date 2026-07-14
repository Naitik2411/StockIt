package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/server"
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
