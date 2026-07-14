package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Stock struct {
	Ticker         string          `json:"ticker"`
	Name           string          `json:"name"`
	CurrentPrice   decimal.Decimal `json:"current_price"`
	PriceChangePct decimal.Decimal `json:"price_change_pct"`
	LastSyncedAt   time.Time       `json:"last_synced_at"`
}
