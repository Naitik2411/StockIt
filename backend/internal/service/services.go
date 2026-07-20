package service

import (
	"github.com/Naitik2411/stockit/internal/lib/job"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
)

type Services struct {
	Auth        *AuthService
	Stock       *StockService
	Portfolio   *PortfolioService
	Leaderboard *LeaderboardService
	Job         *job.JobService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	portfolioService := NewPortfolioService(
		s,
		repos.Portfolio,
		repos.Position,
		repos.Transaction,
	)

	return &Services{
		Job:       s.Job,
		Auth:      NewAuthService(s, repos.User, repos.Portfolio),
		Stock:     NewStockService(s),
		Portfolio: portfolioService,
		Leaderboard: NewLeaderboardService(
			s,
			repos.Portfolio,
			repos.Position,
			repos.User,
		),
	}, nil
}
