package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LeaderboardEntry struct {
	Rank        int             `json:"rank"`
	UserID      uuid.UUID       `json:"user_id"`
	Username    *string         `json:"username"`
	TotalValue  decimal.Decimal `json:"total_value"`
	ReturnPct   decimal.Decimal `json:"return_pct"`
	CashBalance decimal.Decimal `json:"cash_balance"`
}

type LeaderboardPage struct {
	Entries []LeaderboardEntry `json:"entries"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int                `json:"total"`
}

type MyRank struct {
	Rank       int              `json:"rank"`
	Total      int              `json:"total"`
	Percentile decimal.Decimal  `json:"percentile"`
	Entry      LeaderboardEntry `json:"entry"`
}
