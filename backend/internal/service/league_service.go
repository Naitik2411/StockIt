package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/Naitik2411/stockit/internal/lib"
	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type LeagueService struct {
	server        *server.Server
	leagueRepo    *repository.LeagueRepository
	seasonRepo    *repository.SeasonRepository
	portfolioRepo *repository.PortfolioRepository
	userRepo      *repository.UserRepository
}

func NewLeagueService(s *server.Server, leagueRepo *repository.LeagueRepository, seasonRepo *repository.SeasonRepository, portfolioRepo *repository.PortfolioRepository, userRepo *repository.UserRepository) *LeagueService {
	return &LeagueService{
		server:        s,
		leagueRepo:    leagueRepo,
		seasonRepo:    seasonRepo,
		portfolioRepo: portfolioRepo,
		userRepo:      userRepo,
	}
}

type CreateLeagueInput struct {
	Name       string
	MaxMembers int
	SeasonDays int
}

func (s *LeagueService) Create(ctx context.Context, userID uuid.UUID, in CreateLeagueInput) (*model.CreateLeagueResult, error) {
	name := strings.TrimSpace(in.Name)
	nameLen := utf8.RuneCountInString(name)
	if nameLen < 3 || nameLen > 100 {
		code := "INVALID_LEAGUE_NAME"
		return nil, errorss.NewBadRequestError("name must be 3–100 characters", false, &code, nil, nil)
	}

	maxMembers := in.MaxMembers
	if maxMembers == 0 {
		maxMembers = s.server.Config.Integration.MaxLeagueMembers
	}
	if maxMembers == 0 {
		maxMembers = 50
	}
	hardCap := s.server.Config.Integration.MaxLeagueMembers
	if hardCap <= 0 {
		hardCap = 50
	}

	if maxMembers < 2 || maxMembers > hardCap {
		code := "INVALID_MAX_MEMBERS"
		return nil, errorss.NewBadRequestError(
			fmt.Sprintf("maxMembers must be between 2 and %d", hardCap),
			false, &code, nil, nil,
		)
	}

	seasonDays := in.SeasonDays
	if seasonDays == 0 {
		seasonDays = s.server.Config.Integration.DefaultSeasonDays
	}
	if seasonDays == 0 {
		seasonDays = 30
	}
	if seasonDays < 1 || seasonDays > 365 {
		code := "INVALID_SEASON_DAYS"
		return nil, errorss.NewBadRequestError("seasonDays must be between 1 and 365", false, &code, nil, nil)
	}

	inviteCode, err := lib.GenerateInviteCode()
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

	league, err := s.leagueRepo.Create(ctx, tx, name, userID, inviteCode, maxMembers)
	if err != nil {
		return nil, err
	}

	if _, err := s.leagueRepo.AddMember(ctx, tx, league.ID, userID, "admin"); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	season, err := s.seasonRepo.Create(ctx, tx, league.ID, 1, now, now.AddDate(0, 0, seasonDays))
	if err != nil {
		return nil, err
	}

	if _, err := s.portfolioRepo.CreateLeaguePortfolio(ctx, tx, userID, league.ID, season.ID, starting); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.server.Logger.Info().
		Str("operation", "league_create").
		Str("league_id", league.ID.String()).
		Str("user_id", userID.String()).
		Str("invite_code", league.InviteCode).
		Msg("league created")

	return &model.CreateLeagueResult{League: *league, Season: *season}, nil
}

func (s *LeagueService) Join(ctx context.Context, userID uuid.UUID, inviteCode string) (*model.League, error) {
	code := strings.ToUpper(strings.TrimSpace(inviteCode))
	if code == "" {
		c := "INVALID_INVITE_CODE"
		return nil, errorss.NewBadRequestError("invite code is required", false, &c, nil, nil)
	}

	league, err := s.leagueRepo.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if league == nil {
		c := "INVALID_INVITE_CODE"
		return nil, errorss.NewNotFoundError("invalid invite code", false, &c)
	}

	if league.Status != "active" {
		c := "LEAGUE_ARCHIVED"
		return nil, errorss.NewBadRequestError("league is not active", false, &c, nil, nil)
	}

	count, err := s.leagueRepo.CountMembers(ctx, league.ID)
	if err != nil {
		return nil, err
	}

	if count >= league.MaxMembers {
		c := "LEAGUE_FULL"
		return nil, errorss.NewBadRequestError("league is full", false, &c, nil, nil)
	}

	already, err := s.leagueRepo.IsMember(ctx, league.ID, userID)
	if err != nil {
		return nil, err
	}

	if already {
		c := "ALREADY_LEAGUE_MEMBER"
		return nil, errorss.NewBadRequestError("already a member of this league", false, &c, nil, nil)
	}

	season, err := s.seasonRepo.GetCurrentActive(ctx, league.ID)
	if err != nil {
		return nil, err
	}
	if season == nil {
		c := "NO_ACTIVE_SEASON"
		return nil, errorss.NewBadRequestError("league has no active season", false, &c, nil, nil)
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
	if _, err := s.leagueRepo.AddMember(ctx, tx, league.ID, userID, "member"); err != nil {
		return nil, err
	}
	if _, err := s.portfolioRepo.CreateLeaguePortfolio(ctx, tx, userID, league.ID, season.ID, starting); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.server.Logger.Info().
		Str("operation", "league_join").
		Str("league_id", league.ID.String()).
		Str("user_id", userID.String()).
		Msg("joined league")
	return league, nil
}

func (s *LeagueService) ListMine(ctx context.Context, userID uuid.UUID) ([]model.League, error) {
	return s.leagueRepo.ListByUserID(ctx, userID)
}

func (s *LeagueService) GetByID(ctx context.Context, userID uuid.UUID, leagueID uuid.UUID) (*model.LeagueDetail, error) {
	member, err := s.leagueRepo.IsMember(ctx, leagueID, userID)
	if err != nil {
		return nil, err
	}

	if !member {
		c := "LEAGUE_NOT_FOUND"
		return nil, errorss.NewNotFoundError("league not found", false, &c)
	}

	league, err := s.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	if league == nil {
		c := "LEAGUE_NOT_FOUND"
		return nil, errorss.NewNotFoundError("league not found", false, &c)
	}

	count, err := s.leagueRepo.CountMembers(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	season, err := s.seasonRepo.GetCurrentActive(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	return &model.LeagueDetail{
		League:        *league,
		MemberCount:   count,
		CurrentSeason: season,
	}, nil
}

func (s *LeagueService) GetMembers(ctx context.Context, userID, leagueID uuid.UUID) ([]model.LeagueMemberView, error) {
	member, err := s.leagueRepo.IsMember(ctx, leagueID, userID)
	if err != nil {
		return nil, err
	}
	if !member {
		c := "LEAGUE_NOT_FOUND"
		return nil, errorss.NewNotFoundError("league not found", false, &c)
	}
	members, err := s.leagueRepo.ListMembers(ctx, leagueID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	users, err := s.userRepo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]model.LeagueMemberView, 0, len(members))
	for _, m := range members {
		view := model.LeagueMemberView{
			UserID:   m.UserID,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		}
		if u, ok := users[m.UserID]; ok {
			view.Username = u.Username
		}
		out = append(out, view)
	}
	return out, nil
}

func (s *LeagueService) GetMemberRole(ctx context.Context, leagueID, userID uuid.UUID) (string, error) {
	return s.leagueRepo.GetMemberRole(ctx, leagueID, userID)
}

func (s *LeagueService) KickMember(ctx context.Context, adminID, leagueID, targetID uuid.UUID) error {
	if adminID == targetID {
		c := "CANNOT_KICK_ADMIN"
		return errorss.NewBadRequestError("admin cannot kick themselves", false, &c, nil, nil)
	}

	role, err := s.leagueRepo.GetMemberRole(ctx, leagueID, targetID)
	if err != nil {
		return err
	}

	if role == "" {
		c := "MEMBER_NOT_FOUND"
		return errorss.NewBadRequestError("member not found in this league", false, &c, nil, nil)
	}

	if role == "admin" {
		c := "CANNOT_KICK_ADMIN"
		return errorss.NewBadRequestError("cannot kick a league admin", false, &c, nil, nil)
	}

	tx, err := s.server.DB.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.leagueRepo.RemoveMember(ctx, tx, leagueID, targetID); err != nil {
		return err
	}

	if err := s.portfolioRepo.DeleteLeaguePortfolios(ctx, tx, targetID, leagueID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.server.Logger.Info().
		Str("operation", "league_kick_member").
		Str("league_id", leagueID.String()).
		Str("admin_id", adminID.String()).
		Str("target_user_id", targetID.String()).
		Msg("member kicked from league")

	return nil

}

func (s *LeagueService) RegenerateInviteCode(ctx context.Context, leagueID uuid.UUID) (*model.League, error) {
	const maxAttempts = 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := lib.GenerateInviteCode()
		if err != nil {
			return nil, err
		}
		league, err := s.leagueRepo.UpdateInviteCode(ctx, leagueID, code)
		if err == nil {
			s.server.Logger.Info().
				Str("operation", "league_regenerate_invite").
				Str("league_id", leagueID.String()).
				Str("invite_code", league.InviteCode).
				Msg("invite code regenerated")
			return league, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}
		return nil, err
	}
	return nil, errorss.NewInternalServerError()

}
