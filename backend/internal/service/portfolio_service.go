package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/Naitik2411/stockit/internal/lib"
	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

func (s *PortfolioService) GetOrCreatePortfolio(ctx context.Context, userID uuid.UUID) (*model.Portfolio, error) {
	p, err := s.portfolioRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}
	starting := s.server.Config.Integration.StartingBalance
	if starting <= 0 {
		starting = 100000
	}
	return s.portfolioRepo.Create(ctx, userID, starting)
}

func (s *PortfolioService) Buy(ctx context.Context, userID uuid.UUID, ticker string, shareStr string) error {
	ticker = strings.ToUpper(ticker)
	shares, err := decimal.NewFromString(shareStr)
	if err != nil || !shares.IsPositive() {
		code := "INVALID_SHARES"
		return errorss.NewBadRequestError("shares must be a positive number", false, &code, nil, nil)
	}

	cached, err := s.server.Cache.GetPrice(ctx, ticker)
	if err != nil {
		if errors.Is(err, errorss.ErrTickerNotFound) {
			code := "TICKER_NOT_FOUND"
			return errorss.NewNotFoundError("ticker not found or not yet synced", false, &code)
		}
		return err
	}
	price, err := decimal.NewFromString(cached.Price)
	if err != nil {
		return fmt.Errorf("parse price : %w", err)
	}

	tz := s.server.Config.Integration.MarketTimezone
	if tz == "" {
		tz = "America/New_York"
	}
	if !lib.IsMarketOpen(time.Now(), tz) {
		code := "MARKET_CLOSED"
		return errorss.NewBadRequestError("market is closed", false, &code, nil, nil)
	}

	portfolio, err := s.GetOrCreatePortfolio(ctx, userID)
	if err != nil {
		return err
	}
	totalCost := shares.Mul(price)
	if portfolio.CashBalance.LessThan(totalCost) {
		code := "INSUFFICIENT_FUNDS"
		return errorss.NewBadRequestError("not enough funds", false, &code, nil, nil)
	}

	tx, err := s.server.DB.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	newCash := portfolio.CashBalance.Sub(totalCost)
	if err := s.portfolioRepo.UpdateCash(ctx, tx, portfolio.ID, newCash); err != nil {
		return err
	}

	if err := s.positionRepo.UpsertBuy(ctx, tx, portfolio.ID, ticker, shares, price); err != nil {
		return err
	}

	if err := s.txnRepo.Create(ctx, tx, portfolio.ID, ticker, "BUY", shares, price, totalCost); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PortfolioService) Sell(ctx context.Context, userID uuid.UUID, ticker string, shareStr string) error {
	ticker = strings.ToUpper(ticker)
	shares, err := decimal.NewFromString(shareStr)
	if err != nil || !shares.IsPositive() {
		code := "INVALID_SHARES"
		return errorss.NewBadRequestError("shares must be a positive value", false, &code, nil, nil)
	}

	cached, err := s.server.Cache.GetPrice(ctx, ticker)
	if err != nil {
		if errors.Is(err, errorss.ErrTickerNotFound) {
			code := "TICKER_NOT_FOUND"
			return errorss.NewNotFoundError("ticker not found or not yet synced", false, &code)
		}
		return err
	}

	price, err := decimal.NewFromString(cached.Price)
	if err != nil {
		return fmt.Errorf("price parse error : %w", err)
	}

	portfolio, err := s.GetOrCreatePortfolio(ctx, userID)
	if err != nil {
		return err
	}

	tx, err := s.server.DB.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	position, err := s.positionRepo.GetByPortfolioAndTicker(ctx, tx, portfolio.ID, ticker)
	if err != nil {
		return err
	}
	if position == nil {
		code := "POSITION_NOT_FOUND"
		return errorss.NewNotFoundError("position not found", false, &code)
	}
	if position.Shares.LessThan(shares) {
		code := "INSUFFICIENT_SHARES"
		return errorss.NewBadRequestError("not enough shares", false, &code, nil, nil)
	}

	proceeds := shares.Mul(price)
	newCash := portfolio.CashBalance.Add(proceeds)
	if err := s.portfolioRepo.UpdateCash(ctx, tx, portfolio.ID, newCash); err != nil {
		return err
	}
	remaining := position.Shares.Sub(shares)
	if remaining.IsZero() {
		if err := s.positionRepo.Delete(ctx, tx, position.ID); err != nil {
			return err
		}
	} else {
		if err := s.positionRepo.UpdateShares(ctx, tx, position.ID, remaining); err != nil {
			return err
		}
	}
	if err := s.txnRepo.Create(ctx, tx, portfolio.ID, ticker, "SELL", shares, price, proceeds); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PortfolioService) Summary(ctx context.Context, userID uuid.UUID) (*model.PortfolioSummary, error) {
	portfolio, err := s.GetOrCreatePortfolio(ctx, userID)
	if err != nil {
		return nil, err
	}

	positions, err := s.positionRepo.ListByPortfolio(ctx, portfolio.ID)
	if err != nil {
		return nil, err
	}

	starting := decimal.NewFromInt(int64(s.server.Config.Integration.StartingBalance))
	if starting.IsZero() {
		starting = decimal.NewFromInt(100000)
	}

	var withPnL []model.PositionWithPnL
	invested := decimal.Zero
	for _, pos := range positions {
		cached, err := s.server.Cache.GetPrice(ctx, pos.Ticker)
		price := decimal.Zero
		if err == nil {
			price, _ = decimal.NewFromString(cached.Price)
		}
		mkt := price.Mul(pos.Shares)
		pnl := mkt.Sub(pos.AvgBuyPrice.Mul(pos.Shares))
		invested = invested.Add(mkt)
		withPnL = append(withPnL, model.PositionWithPnL{
			Position:      pos,
			CurrentPrice:  price,
			MarketValue:   mkt,
			UnrealizedPnL: pnl,
		})
	}
	totalValue := portfolio.CashBalance.Add(invested)
	returnPct := totalValue.Sub(starting).Div(starting).Mul(decimal.NewFromInt(100))
	return &model.PortfolioSummary{
		Portfolio:  *portfolio,
		Positions:  withPnL,
		TotalValue: totalValue,
		ReturnPct:  returnPct,
	}, nil
}

func (s *PortfolioService) Positions(ctx context.Context, userID uuid.UUID) ([]model.PositionWithPnL, error) {
	summary, err := s.Summary(ctx, userID)
	if err != nil {
		return nil, err
	}
	return summary.Positions, nil
}

func (s *PortfolioService) Transactions(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.Transaction, error) {
	portfolio, err := s.GetOrCreatePortfolio(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.txnRepo.ListByPortfolio(ctx, portfolio.ID, page, limit)
}
