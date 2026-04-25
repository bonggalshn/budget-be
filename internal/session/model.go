package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	TokenHash      string     `json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	InvalidatedAt  *time.Time `json:"invalidated_at,omitempty"`
	LastActivityAt time.Time  `json:"last_activity_at"`
}

func (s *Session) IsActive() bool {
	return s.InvalidatedAt == nil && time.Now().Before(s.ExpiresAt)
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
