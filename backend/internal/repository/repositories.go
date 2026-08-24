package repository

import "github.com/Naitik2411/stockit/internal/server"

type Repositories struct {
	User        *UserRepository
	Stock       *StockRepository
	Portfolio   *PortfolioRepository
	Position    *PositionRepository
	Transaction *TransactionRepository
	League      *LeagueRepository
	Season      *SeasonRepository
	Snapshot    *SnapshotRepository
	ELO         *ELORepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		User:        NewUserRepository(s),
		Stock:       NewStockRepository(s),
		Portfolio:   NewPortfolioRepository(s),
		Position:    NewPositionRepository(s),
		Transaction: NewTransactionRepository(s),
		League:      NewLeagueRepository(s),
		Season:      NewSeasonRepository(s),
		Snapshot:    NewSnapshotRepository(s),
		ELO:         NewELORepository(s),
	}
}
