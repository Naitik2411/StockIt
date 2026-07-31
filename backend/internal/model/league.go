package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type League struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	CreatedBy  uuid.UUID `json:"created_by"`
	InviteCode string    `json:"invite_code"`
	MaxMembers int       `json:"max_members"`
	IsPublic   bool      `json:"is_public"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type LeagueMember struct {
	ID       uuid.UUID `json:"id"`
	LeagueID uuid.UUID `json:"league_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"` // admin | member
	JoinedAt time.Time `json:"joined_at"`
}

type Season struct {
	ID           uuid.UUID `json:"id"`
	LeagueID     uuid.UUID `json:"league_id"`
	SeasonNumber int       `json:"season_number"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Status       string    `json:"status"` // active | completed | pending
	CreatedAt    time.Time `json:"created_at"`
}

type LeagueMemberView struct {
	UserID   uuid.UUID `json:"user_id"`
	Username *string   `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type LeagueDetail struct {
	League
	MemberCount   int     `json:"member_count"`
	CurrentSeason *Season `json:"current_season,omitempty"`
}

type CreateLeagueResult struct {
	League League `json:"league"`
	Season Season `json:"season"`
}

type LeagueStanding struct {
	UserID     uuid.UUID       `json:"user_id"`
	Username   *string         `json:"username,omitempty"`
	TotalValue decimal.Decimal `json:"total_value"`
	ReturnPct  decimal.Decimal `json:"return_pct"`
	Rank       int             `json:"rank"`
}
