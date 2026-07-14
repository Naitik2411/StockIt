package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID           uuid.UUID       `json:"id"`
	PortfolioID  uuid.UUID       `json:"portfolio_id"`
	Ticker       string          `json:"ticker"`
	Type         string          `json:"type"`
	Shares       decimal.Decimal `json:"shares"`
	PriceAtTrade decimal.Decimal `json:"price_at_trade"`
	TotalValue   decimal.Decimal `json:"total_value"`
	CreatedAt    time.Time       `json:"created_at"`
}
