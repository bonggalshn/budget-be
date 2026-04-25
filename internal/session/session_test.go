package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSession_IsActive(t *testing.T) {
	t.Run("returns true when session is not invalidated and not expired", func(t *testing.T) {
		s := &Session{
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		if !s.IsActive() {
			t.Error("expected active session")
		}
	})

	t.Run("returns false when session is expired", func(t *testing.T) {
		s := &Session{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		if s.IsActive() {
			t.Error("expected expired session to be inactive")
		}
	})

	t.Run("returns false when session is invalidated", func(t *testing.T) {
		now := time.Now()
		s := &Session{
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			InvalidatedAt: &now,
		}

		if s.IsActive() {
			t.Error("expected invalidated session to be inactive")
		}
	})
}

func TestSession_IsExpired(t *testing.T) {
	t.Run("returns true when current time is past expires_at", func(t *testing.T) {
		s := &Session{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		if !s.IsExpired() {
			t.Error("expected session to be expired")
		}
	})

	t.Run("returns false when current time is before expires_at", func(t *testing.T) {
		s := &Session{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		if s.IsExpired() {
			t.Error("expected session to not be expired")
		}
	})
}

func TestSession_Model(t *testing.T) {
	t.Run("creates session with ID", func(t *testing.T) {
		s := &Session{
			ID:         uuid.New(),
			UserID:     uuid.New(),
			TokenHash:  "abc123",
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}

		if s.ID == uuid.Nil {
			t.Error("expected session ID to be set")
		}
	})
}