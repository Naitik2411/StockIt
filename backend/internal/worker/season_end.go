package worker

import (
	"context"
	"sync"
	"time"

	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
)

const maxConcurrentCloses = 3

type SeasonEndWorker struct {
	server     *server.Server
	seasonRepo *repository.SeasonRepository
	seasonSvc  *service.SeasonService
	interval   time.Duration
	log        zerolog.Logger
}

func NewSeasonEndWorker(
	s *server.Server,
	seasonRepo *repository.SeasonRepository,
	seasonSvc *service.SeasonService,
	intervalSecs int,
) *SeasonEndWorker {
	if intervalSecs <= 0 {
		intervalSecs = 300
	}
	return &SeasonEndWorker{
		server:     s,
		seasonRepo: seasonRepo,
		seasonSvc:  seasonSvc,
		interval:   time.Duration(intervalSecs) * time.Second,
		log:        s.Logger.With().Str("component", "season_end").Logger(),
	}
}

func (w *SeasonEndWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.CheckAndClose(ctx)

	for {
		select {
		case <-ctx.Done():
			w.log.Info().Msg("season end worker shutting down")
			return
		case <-ticker.C:
			w.CheckAndClose(ctx)
		}
	}

}

func (w *SeasonEndWorker) CheckAndClose(ctx context.Context) {
	if w.server.LoggerService != nil {
		if app := w.server.LoggerService.GetApplication(); app != nil {
			txn := app.StartTransaction("SeasonEndWorker/CheckAnClose")
			defer txn.End()
			ctx = newrelic.NewContext(ctx, txn)
		}
	}
	expired, err := w.seasonRepo.FindExpired(ctx)
	if err != nil {
		w.log.Error().Err(err).Msg("finding expired seasons")
		return
	}

	if len(expired) == 0 {
		return
	}

	w.log.Info().Int("count", len(expired)).Msg("closing expired seasons")

	sem := make(chan struct{}, maxConcurrentCloses)
	var wg sync.WaitGroup

	for _, season := range expired {
		wg.Add(1)
		go func(seasonID string, id [16]byte) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					w.log.Error().Interface("panic", r).Str("season_id", seasonID).Msg("panic during season close")
				}
			}()

			sem <- struct{}{}
			defer func() { <-sem }()
			start := time.Now()
			if err := w.seasonSvc.Close(ctx, id); err != nil {
				w.log.Error().Err(err).Str("season_id", seasonID).Msg("failed to close season")
				return
			}
			w.log.Info().Str("season_id", seasonID).Dur("duration", time.Since(start)).Msg("season closed")
		}(season.ID.String(), season.ID)
	}
	wg.Wait()
}
