package service

import (
	"context"

	"github.com/Naitik2411/stockit/internal/lib"
	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type ELOService struct {
	server       *server.Server
	eloRepo      *repository.ELORepository
	snapshotRepo *repository.SnapshotRepository
	userRepo     *repository.UserRepository
}

func NewELOService(
	s *server.Server,
	eloRepo *repository.ELORepository,
	snapshotRepo *repository.SnapshotRepository,
	userRepo *repository.UserRepository,
) *ELOService {
	return &ELOService{
		server:       s,
		eloRepo:      eloRepo,
		snapshotRepo: snapshotRepo,
		userRepo:     userRepo,
	}
}

func (s *ELOService) kFactor() int {
	k := s.server.Config.Integration.ELOKFactor
	if k <= 0 {
		k = lib.DefaultELOK
	}
	return k
}

func (s *ELOService) startingRating() int {
	r := s.server.Config.Integration.ELOStartingRating
	if r <= 0 {
		r = 1000
	}
	return r
}

// UpdateRatings runs inside the season-close transaction: read current ratings,
// compute new ones with pure math, write them back.
func (s *ELOService) UpdateRatings(ctx context.Context, tx pgx.Tx, standings []lib.Standing) error {
	if len(standings) == 0 {
		return nil
	}

	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("elo-update").End()
	}

	userIDs := make([]uuid.UUID, 0, len(standings))
	for _, st := range standings {
		userIDs = append(userIDs, st.UserID)
	}

	current, err := s.eloRepo.GetRatings(ctx, tx, userIDs)
	if err != nil {
		return err
	}

	starting := s.startingRating()
	newRatings := lib.CalcSeasonELOChanges(current, standings, s.kFactor(), starting)

	updates := make([]repository.RatingUpdate, 0, len(newRatings))
	for _, st := range standings {
		updates = append(updates, repository.RatingUpdate{
			UserID: st.UserID,
			Rating: newRatings[st.UserID],
		})
	}

	if err := s.eloRepo.UpsertBulk(ctx, tx, updates, starting); err != nil {
		return err
	}

	s.server.Logger.Info().
		Str("operation", "elo_update").
		Int("players", len(updates)).
		Int("k_factor", s.kFactor()).
		Msg("elo ratings updated")
	return nil
}

func (s *ELOService) GlobalRankings(ctx context.Context, page, limit int) (*model.ELORankingPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	ratings, total, err := s.eloRepo.ListTop(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uuid.UUID, 0, len(ratings))
	for _, r := range ratings {
		userIDs = append(userIDs, r.UserID)
	}
	users, err := s.userRepo.ListByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	entries := make([]model.ELORankingEntry, 0, len(ratings))
	for i, r := range ratings {
		entry := model.ELORankingEntry{
			ELORating: r,
			Rank:      (page-1)*limit + i + 1,
		}
		if u, ok := users[r.UserID]; ok {
			entry.Username = u.Username
		}
		entries = append(entries, entry)
	}

	return &model.ELORankingPage{
		Entries: entries,
		Page:    page,
		Limit:   limit,
		Total:   total,
	}, nil
}

// UserRating returns a user's rating plus their season-by-season history.
// Users who have never finished a season get a synthetic starting record.
func (s *ELOService) UserRating(ctx context.Context, userID uuid.UUID) (*model.UserELODetail, error) {
	rating, err := s.eloRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if rating == nil {
		starting := s.startingRating()
		rating = &model.ELORating{
			UserID:        userID,
			Rating:        starting,
			SeasonsPlayed: 0,
			PeakRating:    starting,
		}
	}

	history, err := s.snapshotRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	detail := &model.UserELODetail{
		ELORating: *rating,
		History:   history,
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		detail.Username = user.Username
	}
	return detail, nil
}
