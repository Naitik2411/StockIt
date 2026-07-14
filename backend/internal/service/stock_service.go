package service

import (
	"context"
	"errors"

	"github.com/Naitik2411/stockit/internal/cache"
	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/Naitik2411/stockit/internal/server"
)

type StockService struct {
	server *server.Server
}

func NewStockService(s *server.Server) *StockService {
	return &StockService{
		server: s,
	}
}

func (s *StockService) GetTicker(ctx context.Context, ticker string) (cache.StockPrice, error) {
	price, err := s.server.Cache.GetPrice(ctx, ticker)
	if err != nil {
		if errors.Is(err, errorss.ErrTickerNotFound) {
			return cache.StockPrice{}, errorss.NewNotFoundError("ticker not found or not yet synced", false, strPtr("TICKER_NOT_FOUND"))
		}
		return cache.StockPrice{}, err
	}
	return price, nil
}

func strPtr(s string) *string { return &s }
