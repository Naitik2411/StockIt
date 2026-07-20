package worker

import (
	"context"
	"sync"
	"time"

	"github.com/Naitik2411/stockit/internal/apiclient"
	"github.com/Naitik2411/stockit/internal/cache"
	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

type PriceSyncWorker struct {
	server    *server.Server
	client    apiclient.PriceSource
	stockRepo *repository.StockRepository
	watchlist []string
	interval  time.Duration
	log       zerolog.Logger
}

func NewPriceSyncWorker(
	s *server.Server,
	client apiclient.PriceSource,
	stockRepo *repository.StockRepository,
	watchlist []string,
	intervalSecs int,
) *PriceSyncWorker {
	if intervalSecs <= 0 {
		intervalSecs = 60
	}
	return &PriceSyncWorker{
		server:    s,
		client:    client,
		stockRepo: stockRepo,
		watchlist: watchlist,
		interval:  time.Duration(intervalSecs) * time.Second,
		log:       s.Logger.With().Str("component", "price_sync").Logger(),
	}
}

func (w *PriceSyncWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.log.Info().Msg("price sync worker started")
	w.sync(ctx)

	for {
		select {
		case <-ctx.Done():
			w.log.Info().Msg("price sync worker shutting down")
			return
		case <-ticker.C:
			w.sync(ctx)
		}
	}
}

func (w *PriceSyncWorker) sync(ctx context.Context) {
	if w.server.LoggerService != nil {
		if app := w.server.LoggerService.GetApplication(); app != nil {
			txn := app.StartTransaction("PriceSyncWorker/sync")
			defer txn.End()
			ctx = newrelic.NewContext(ctx, txn)
		}
	}

	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for _, t := range w.watchlist {
		wg.Add(1)
		go func(ticker string) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					w.log.Error().Interface("panic", r).Str("ticker", ticker).Msg("panic during price sync")
				}
			}()
			sem <- struct{}{}
			defer func() {
				<-sem
			}()
			w.syncTicker(ctx, ticker)
		}(t)
	}
	wg.Wait()
}

func (w *PriceSyncWorker) syncTicker(ctx context.Context, ticker string) {
	start := time.Now()
	if txn := newrelic.FromContext(ctx); txn != nil {
		segment := txn.StartSegment("price-sync")
		defer segment.End()
	}

	price, changePct, err := w.client.GetQuote(ctx, ticker)
	if err != nil {
		w.log.Error().Err(err).Str("ticker", ticker).Msg("failed to fetch price")
		return
	}
	now := time.Now().UTC()

	err = w.server.Cache.SetPrice(ctx, ticker, cache.StockPrice{
		Ticker:    ticker,
		Price:     price,
		ChangePct: changePct,
		SyncedAt:  now,
	})
	if err != nil {
		w.log.Error().Err(err).Str("ticker", ticker).Msg("failed to cache price")
		return
	}

	priceDec, _ := decimal.NewFromString(price)
	changeDec, _ := decimal.NewFromString(changePct)

	if err := w.stockRepo.Upsert(ctx, &model.Stock{
		Ticker:         ticker,
		Name:           ticker,
		CurrentPrice:   priceDec,
		PriceChangePct: changeDec,
		LastSyncedAt:   now,
	}); err != nil {
		w.log.Error().Err(err).Str("ticker", ticker).Msg("failed to upsert stock")
		return
	}

	w.log.Info().
		Str("ticker", ticker).
		Str("price", price).
		Dur("latency", time.Since(start)).
		Msg("price synced")
}
