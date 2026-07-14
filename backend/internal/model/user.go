package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	ClerkUserID string    `json:"clerk_user_id"`
	Username    *string   `json:"username,omitempty"`
	Email       *string   `json:"email,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
