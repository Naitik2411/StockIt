package service

import (
	"github.com/Naitik2411/go-production/internal/lib/job"
	"github.com/Naitik2411/go-production/internal/repository"
	"github.com/Naitik2411/go-production/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	return &Services{
		Job:  s.Job,
		Auth: authService,
	}, nil
}
