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
	server        *server.Server
	repo          *repository.UserRepository
	portfolioRepo *repository.PortfolioRepository
}

func NewAuthService(s *server.Server, repo *repository.UserRepository, portfolioRepo *repository.PortfolioRepository) *AuthService {
	clerk.SetKey(s.Config.Auth.SecretKey)
	return &AuthService{
		server:        s,
		repo:          repo,
		portfolioRepo: portfolioRepo,
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

	starting := a.server.Config.Integration.StartingBalance
	if starting <= 0 {
		starting = 100000
	}

	if _, err := a.portfolioRepo.Create(ctx, user.ID, starting); err != nil {
		return nil, fmt.Errorf("create portfolio: %w", err)
	}

	a.server.Logger.Info().
		Str("clerk_user_id", clerkUserID).
		Str("user_id", user.ID.String()).
		Msg("created new user from clerk identity")
	return user, nil
}
