package repository

import "github.com/Naitik2411/stockit/internal/server"

type Repositories struct {
	User  *UserRepository
	Stock *StockRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		User:  NewUserRepository(s),
		Stock: NewStockRepository(s),
	}
}
