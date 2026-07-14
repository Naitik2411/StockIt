package service

import (
	"context"
	"fmt"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"

	"github.com/clerk/clerk-sdk-go/v2"
)

type AuthService struct {
	server *server.Server
	repo   *repository.UserRepository
}

func NewAuthService(s *server.Server, repo *repository.UserRepository) *AuthService {
	clerk.SetKey(s.Config.Auth.SecretKey)
	return &AuthService{
		server: s,
		repo:   repo,
	}
}

func (a *AuthService) CreateOrGetUser(ctx context.Context, clerkUserID string) (*model.User, error) {
	user, err := a.repo.GetByClerkID(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}
	user, err = a.repo.Create(ctx, clerkUserID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	a.server.Logger.Info().
		Str("clerk_user_id", clerkUserID).
		Str("user_id", user.ID.String()).
		Msg("created new user from clerk identity")
	return user, nil
}
