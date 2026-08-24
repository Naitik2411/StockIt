package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/Naitik2411/stockit/internal/lib"
	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/shopspring/decimal"
)

type SeasonService struct {
	server        *server.Server
	seasonRepo    *repository.SeasonRepository
	snapshotRepo  *repository.SnapshotRepository
	leagueRepo    *repository.LeagueRepository
	portfolioRepo *repository.PortfolioRepository
	positionRepo  *repository.PositionRepository
	userRepo      *repository.UserRepository
	eloService    *ELOService
}

func NewSeasonService(
	s *server.Server,
	seasonRepo *repository.SeasonRepository,
	snapshotRepo *repository.SnapshotRepository,
	leagueRepo *repository.LeagueRepository,
	portfolioRepo *repository.PortfolioRepository,
	positionRepo *repository.PositionRepository,
	userRepo *repository.UserRepository,
	eloService *ELOService,
) *SeasonService {
	return &SeasonService{
		server:        s,
		seasonRepo:    seasonRepo,
		snapshotRepo:  snapshotRepo,
		leagueRepo:    leagueRepo,
		portfolioRepo: portfolioRepo,
		positionRepo:  positionRepo,
		userRepo:      userRepo,
		eloService:    eloService,
	}
}

func (s *SeasonService) startingBalance() decimal.Decimal {
	starting := s.server.Config.Integration.StartingBalance
	if starting <= 0 {
		starting = 100000
	}
	return decimal.NewFromInt(int64(starting))
}

func (s *SeasonService) requireMember(ctx context.Context, leagueID, userID uuid.UUID) error {
	member, err := s.leagueRepo.IsMember(ctx, leagueID, userID)
	if err != nil {
		return err
	}
	if !member {
		c := "LEAGUE_NOT_FOUND"
		return errorss.NewNotFoundError("league not found", false, &c)
	}
	return nil
}

func (s *SeasonService) GetCurrent(ctx context.Context, userID, leagueID uuid.UUID) (*model.CurrentSeason, error) {
	if err := s.requireMember(ctx, leagueID, userID); err != nil {
		return nil, err
	}

	season, err := s.seasonRepo.GetCurrentActive(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	if season == nil {
		c := "NO_ACTIVE_SEASON"
		return nil, errorss.NewNotFoundError("league has no active season", false, &c)
	}

	count, err := s.leagueRepo.CountMembers(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	remaining := time.Until(season.EndDate)
	days := 0
	if remaining > 0 {
		days = int(math.Ceil(remaining.Hours() / 24))
	}

	return &model.CurrentSeason{
		Season:        *season,
		DaysRemaining: days,
		MemberCount:   count,
		IsExpired:     remaining <= 0,
	}, nil
}

func (s *SeasonService) GetHistory(
	ctx context.Context,
	userID, leagueID uuid.UUID,
	page, limit int,
) (*model.SeasonHistoryPage, error) {
	if err := s.requireMember(ctx, leagueID, userID); err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	seasons, total, err := s.seasonRepo.ListByLeague(ctx, leagueID, page, limit)
	if err != nil {
		return nil, err
	}

	seasonIDs := make([]uuid.UUID, 0, len(seasons))
	for _, season := range seasons {
		seasonIDs = append(seasonIDs, season.ID)
	}

	snapshotsBySeason, err := s.snapshotRepo.ListBySeasonIDs(ctx, seasonIDs)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uuid.UUID, 0)
	for _, snaps := range snapshotsBySeason {
		for _, snap := range snaps {
			userIDs = append(userIDs, snap.UserID)
		}
	}
	users, err := s.userRepo.ListByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	entries := make([]model.SeasonHistoryEntry, 0, len(seasons))
	for _, season := range seasons {
		snaps := snapshotsBySeason[season.ID]
		standings := make([]model.SnapshotView, 0, len(snaps))
		for _, snap := range snaps {
			view := model.SnapshotView{SeasonSnapshot: snap}
			if u, ok := users[snap.UserID]; ok {
				view.Username = u.Username
			}
			standings = append(standings, view)
		}
		entries = append(entries, model.SeasonHistoryEntry{
			Season:    season,
			Standings: standings,
		})
	}

	return &model.SeasonHistoryPage{
		Seasons: entries,
		Page:    page,
		Limit:   limit,
		Total:   total,
	}, nil
}

// Close snapshots every member's final standing and marks the season completed.
func (s *SeasonService) Close(ctx context.Context, seasonID uuid.UUID) error {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("season-close").End()
	}

	tx, err := s.server.DB.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	season, err := s.seasonRepo.GetForUpdate(ctx, tx, seasonID)
	if err != nil {
		return err
	}
	if season == nil {
		return fmt.Errorf("close season %s: not found", seasonID)
	}
	if season.Status != "active" {
		// already closed by another pass — nothing to do
		return nil
	}

	portfolios, err := s.portfolioRepo.ListBySeasonRef(ctx, seasonID)
	if err != nil {
		return err
	}

	starting := s.startingBalance()
	rows := make([]repository.SnapshotInput, 0, len(portfolios))

	for _, p := range portfolios {
		positions, err := s.positionRepo.ListByPortfolio(ctx, p.ID)
		if err != nil {
			return fmt.Errorf("list positions for portfolio %s: %w", p.ID, err)
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

		rows = append(rows, repository.SnapshotInput{
			UserID:     p.UserID,
			FinalValue: totalValue,
			ReturnPct:  returnPct,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].ReturnPct.GreaterThan(rows[j].ReturnPct)
	})
	for i := range rows {
		rows[i].Rank = i + 1
	}

	if err := s.snapshotRepo.CreateBulk(ctx, tx, seasonID, season.LeagueID, rows); err != nil {
		return err
	}

	standings := make([]lib.Standing, 0, len(rows))
	for _, row := range rows {
		standings = append(standings, lib.Standing{UserID: row.UserID, Rank: row.Rank})
	}
	if err := s.eloService.UpdateRatings(ctx, tx, standings); err != nil {
		return err
	}

	if err := s.seasonRepo.MarkCompleted(ctx, tx, seasonID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.server.Logger.Info().
		Str("operation", "season_close").
		Str("season_id", seasonID.String()).
		Str("league_id", season.LeagueID.String()).
		Int("season_number", season.SeasonNumber).
		Int("member_count", len(rows)).
		Msg("season closed")
	return nil
}

// StartNext closes the active season (if any) then opens season N+1 with a fresh
// $100k portfolio for every current member.
func (s *SeasonService) StartNext(ctx context.Context, leagueID uuid.UUID, seasonDays int) (*model.Season, error) {
	if seasonDays == 0 {
		seasonDays = s.server.Config.Integration.DefaultSeasonDays
	}
	if seasonDays == 0 {
		seasonDays = 30
	}
	if seasonDays < 1 || seasonDays > 365 {
		c := "INVALID_SEASON_DAYS"
		return nil, errorss.NewBadRequestError("seasonDays must be between 1 and 365", false, &c, nil, nil)
	}

	// close the current season first so snapshots exist before the reset
	current, err := s.seasonRepo.GetCurrentActive(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	if current != nil {
		if err := s.Close(ctx, current.ID); err != nil {
			return nil, err
		}
	}

	members, err := s.leagueRepo.ListMembers(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	starting := s.server.Config.Integration.StartingBalance
	if starting <= 0 {
		starting = 100000
	}

	tx, err := s.server.DB.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	maxNumber, err := s.seasonRepo.MaxSeasonNumber(ctx, tx, leagueID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	season, err := s.seasonRepo.Create(ctx, tx, leagueID, maxNumber+1, now, now.AddDate(0, 0, seasonDays))
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		if _, err := s.portfolioRepo.CreateLeaguePortfolio(ctx, tx, m.UserID, leagueID, season.ID, starting); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.server.Logger.Info().
		Str("operation", "season_start_next").
		Str("league_id", leagueID.String()).
		Str("season_id", season.ID.String()).
		Int("season_number", season.SeasonNumber).
		Int("member_count", len(members)).
		Msg("next season started")
	return season, nil
}
