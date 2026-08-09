package worker

import (
	"context"
	"time"

	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
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

}
