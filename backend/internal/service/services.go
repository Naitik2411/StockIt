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
	League      *LeagueService
	Season      *SeasonService
	ELO         *ELOService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	portfolioService := NewPortfolioService(
		s,
		repos.Portfolio,
		repos.Position,
		repos.Transaction,
	)

	eloService := NewELOService(s, repos.ELO, repos.Snapshot, repos.User)

	seasonService := NewSeasonService(
		s,
		repos.Season,
		repos.Snapshot,
		repos.League,
		repos.Portfolio,
		repos.Position,
		repos.User,
		eloService,
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
			repos.Season,
		),
		League: NewLeagueService(
			s,
			repos.League,
			repos.Season,
			repos.Portfolio,
			repos.User,
		),
		Season: seasonService,
		ELO:    eloService,
	}, nil
}
