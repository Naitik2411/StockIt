package service

import (
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
)

type PortfolioService struct {
	server        *server.Server
	portfolioRepo *repository.PortfolioRepository
	positionRepo  *repository.PositionRepository
	txnRepo       *repository.TransactionRepository
}

func NewPortfolioService(
	s *server.Server,
	portfolioRepo *repository.PortfolioRepository,
	positionRepo *repository.PositionRepository,
	txnRepo *repository.TransactionRepository,
) *PortfolioService {
	return &PortfolioService{
		server:        s,
		portfolioRepo: portfolioRepo,
		positionRepo:  positionRepo,
		txnRepo:       txnRepo,
	}
}
