package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/shopspring/decimal"
)

type LeaderboardService struct {
	server        *server.Server
	portfolioRepo *repository.PortfolioRepository
	positionRepo  *repository.PositionRepository
	userRepo      *repository.UserRepository
	seasonRepo    *repository.SeasonRepository
}

func NewLeaderboardService(
	s *server.Server,
	portfolioRepo *repository.PortfolioRepository,
	positionRepo *repository.PositionRepository,
	userRepo *repository.UserRepository,
	seasonRepo *repository.SeasonRepository,
) *LeaderboardService {
	return &LeaderboardService{
		server:        s,
		portfolioRepo: portfolioRepo,
		positionRepo:  positionRepo,
		userRepo:      userRepo,
		seasonRepo:    seasonRepo,
	}
}

func (s *LeaderboardService) startingBalance() decimal.Decimal {
	starting := decimal.NewFromInt(int64(s.server.Config.Integration.StartingBalance))
	if starting.IsZero() {
		starting = decimal.NewFromInt(100000)
	}
	return starting
}

func (s *LeaderboardService) Global(ctx context.Context, page, limit int) (*model.LeaderboardPage, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 50
	}

	entries, err := s.getOrCompute(ctx)
	if err != nil {
		return nil, err
	}

	total := len(entries)
	start := (page - 1) * limit

	if start > total {
		start = total
	}

	end := start + limit
	if end > total {
		end = total
	}
	return &model.LeaderboardPage{
		Entries: entries[start:end],
		Page:    page,
		Limit:   limit,
		Total:   total,
	}, nil
}

func (s *LeaderboardService) MyRank(ctx context.Context, userID uuid.UUID) (*model.MyRank, error) {
	entries, err := s.getOrCompute(ctx)
	if err != nil {
		return nil, err
	}

	total := len(entries)
	for _, e := range entries {
		if e.UserID == userID {
			pct := decimal.Zero
			if total > 1 {
				pct = decimal.NewFromInt(int64(total - e.Rank)).
					Div(decimal.NewFromInt(int64(total - 1))).
					Mul(decimal.NewFromInt(100))
			}
			return &model.MyRank{
				Rank:       e.Rank,
				Total:      total,
				Percentile: pct,
				Entry:      e,
			}, nil
		}
	}
	return &model.MyRank{
		Rank:       0,
		Total:      total,
		Percentile: decimal.Zero,
		Entry: model.LeaderboardEntry{
			UserID: userID,
		},
	}, nil
}

func (s *LeaderboardService) getOrCompute(ctx context.Context) ([]model.LeaderboardEntry, error) {
	cached, err := s.server.Cache.GetLeaderboard(ctx)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	start := time.Now()
	entries, err := s.compute(ctx)
	if err != nil {
		return nil, err
	}

	s.server.Logger.Info().
		Str("operation", "leaderboard_compute").
		Int("entries", len(entries)).
		Dur("duration", time.Since(start)).
		Msg("leaderboard computed")

	ttl := s.server.Config.Integration.LeaderboardCacheTTL
	if err := s.server.Cache.SetLeaderboard(ctx, entries, ttl); err != nil {
		// log but still return computed result
		s.server.Logger.Error().Err(err).Msg("failed to cache leaderboard")
	}

	return entries, nil
}

func (s *LeaderboardService) compute(ctx context.Context) ([]model.LeaderboardEntry, error) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("leaderboard-compute").End()
	}

	portfolios, err := s.portfolioRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	starting := s.startingBalance()

	userIDs := make([]uuid.UUID, 0, len(portfolios))
	for _, p := range portfolios {
		userIDs = append(userIDs, p.UserID)
	}

	users, err := s.userRepo.ListByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	entries := make([]model.LeaderboardEntry, 0, len(portfolios))
	for _, p := range portfolios {
		positions, err := s.positionRepo.ListByPortfolio(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("list positions for portfolio %s: %w", p.ID, err)
		}

		invested := decimal.Zero
		for _, pos := range positions {
			cached, err := s.server.Cache.GetPrice(ctx, pos.Ticker)
			price := decimal.Zero
			if err == nil {
				price, _ = decimal.NewFromString(cached.Price)
			}
			invested = invested.Add(price.Mul(pos.Shares))
		}

		totalValue := p.CashBalance.Add(invested)
		returnPct := totalValue.Sub(starting).Div(starting).Mul(decimal.NewFromInt(100))
		var username *string
		if u, ok := users[p.UserID]; ok {
			username = u.Username
		}
		entries = append(entries, model.LeaderboardEntry{
			UserID:      p.UserID,
			Username:    username,
			TotalValue:  totalValue,
			ReturnPct:   returnPct,
			CashBalance: p.CashBalance,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ReturnPct.GreaterThan(entries[j].ReturnPct)
	})
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries, nil
}

func (s *LeaderboardService) League(ctx context.Context, leagueID uuid.UUID, page, limit int) (*model.LeaderboardPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	entries, err := s.getOrComputeLeague(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	total := len(entries)
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return &model.LeaderboardPage{
		Entries: entries[start:end],
		Page:    page,
		Limit:   limit,
		Total:   total,
	}, nil
}

func (s *LeaderboardService) LeagueMyRank(ctx context.Context, leagueID, userID uuid.UUID) (*model.MyRank, error) {
	entries, err := s.getOrComputeLeague(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	total := len(entries)
	for _, e := range entries {
		if e.UserID == userID {
			pct := decimal.Zero
			if total > 1 {
				pct = decimal.NewFromInt(int64(total - e.Rank)).
					Div(decimal.NewFromInt(int64(total - 1))).
					Mul(decimal.NewFromInt(100))
			}
			return &model.MyRank{
				Rank:       e.Rank,
				Total:      total,
				Percentile: pct,
				Entry:      e,
			}, nil
		}
	}

	return &model.MyRank{
		Rank:       0,
		Total:      total,
		Percentile: decimal.Zero,
		Entry:      model.LeaderboardEntry{UserID: userID},
	}, nil
}

func (s *LeaderboardService) getOrComputeLeague(ctx context.Context, leagueID uuid.UUID) ([]model.LeaderboardEntry, error) {
	season, err := s.seasonRepo.GetCurrentActive(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	if season == nil {
		code := "NO_ACTIVE_SEASON"
		return nil, errorss.NewNotFoundError("league has no active season", false, &code)
	}

	cached, err := s.server.Cache.GetLeagueLeaderboard(ctx, leagueID, season.ID)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	start := time.Now()
	entries, err := s.computeLeague(ctx, season.ID)
	if err != nil {
		return nil, err
	}

	s.server.Logger.Info().
		Str("operation", "league_leaderboard_compute").
		Str("league_id", leagueID.String()).
		Str("season_id", season.ID.String()).
		Int("entries", len(entries)).
		Dur("duration", time.Since(start)).
		Msg("league leaderboard computed")

	ttl := s.server.Config.Integration.LeaderboardCacheTTL
	if err := s.server.Cache.SetLeagueLeaderboard(ctx, leagueID, season.ID, entries, ttl); err != nil {
		s.server.Logger.Error().Err(err).Msg("failed to cache league leaderboard")
	}

	return entries, nil
}

func (s *LeaderboardService) computeLeague(ctx context.Context, seasonID uuid.UUID) ([]model.LeaderboardEntry, error) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("league-leaderboard-compute").End()
	}

	portfolios, err := s.portfolioRepo.ListBySeasonRef(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	starting := s.startingBalance()

	userIDs := make([]uuid.UUID, 0, len(portfolios))
	for _, p := range portfolios {
		userIDs = append(userIDs, p.UserID)
	}
	users, err := s.userRepo.ListByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	entries := make([]model.LeaderboardEntry, 0, len(portfolios))
	for _, p := range portfolios {
		positions, err := s.positionRepo.ListByPortfolio(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("list positions for portfolio %s: %w", p.ID, err)
		}

		invested := decimal.Zero
		for _, pos := range positions {
			price := decimal.Zero
			if cached, err := s.server.Cache.GetPrice(ctx, pos.Ticker); err == nil {
				price, _ = decimal.NewFromString(cached.Price)
			}
			invested = invested.Add(price.Mul(pos.Shares))
		}

		totalValue := p.CashBalance.Add(invested)
		returnPct := totalValue.Sub(starting).Div(starting).Mul(decimal.NewFromInt(100))

		var username *string
		if u, ok := users[p.UserID]; ok {
			username = u.Username
		}

		entries = append(entries, model.LeaderboardEntry{
			UserID:      p.UserID,
			Username:    username,
			TotalValue:  totalValue,
			ReturnPct:   returnPct,
			CashBalance: p.CashBalance,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ReturnPct.GreaterThan(entries[j].ReturnPct)
	})
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries, nil
}
