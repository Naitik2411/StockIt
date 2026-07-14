package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Portfolio struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	CashBalance decimal.Decimal `json:"cash_balance"`
	SeasonID    string          `json:"season_id"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Position struct {
	ID          uuid.UUID       `json:"id"`
	PortfolioID uuid.UUID       `json:"portfolio_id"`
	Ticker      string          `json:"ticker"`
	Shares      decimal.Decimal `json:"shares"`
	AvgBuyPrice decimal.Decimal `json:"avg_buy_price"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Computed on read — never stored
type PositionWithPnL struct {
	Position
	CurrentPrice  decimal.Decimal `json:"current_price"`
	MarketValue   decimal.Decimal `json:"market_value"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`
}
type PortfolioSummary struct {
	Portfolio
	Positions  []PositionWithPnL `json:"positions"`
	TotalValue decimal.Decimal   `json:"total_value"`
	ReturnPct  decimal.Decimal   `json:"return_pct"`
}
