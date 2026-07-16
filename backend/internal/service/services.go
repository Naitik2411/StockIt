package service

import (
	"github.com/Naitik2411/stockit/internal/lib/job"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
)

type Services struct {
	Auth  *AuthService
	Stock *StockService
	Job   *job.JobService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {

	return &Services{
		Job:   s.Job,
		Auth:  NewAuthService(s, repos.User, repos.Portfolio),
		Stock: NewStockService(s),
	}, nil
}
