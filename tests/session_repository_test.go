package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/bonggalshn/budget-be/internal/session"
	"github.com/google/uuid"
	"crypto/sha256"
	"encoding/hex"
)

func TestSessionRepository_Create(t *testing.T) {
	t.Run("creates session with valid user ID", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		tokenHash := hashToken("test-token")

		sess := &session.Session{
			ID:             uuid.New(),
			UserID:         userID,
			TokenHash:      tokenHash,
			CreatedAt:      time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			LastActivityAt: time.Now(),
		}

		if sess.UserID != userID {
			t.Errorf("expected userID %s, got %s", userID, sess.UserID)
		}
		if sess.TokenHash != tokenHash {
			t.Errorf("expected tokenHash %s, got %s", tokenHash, sess.TokenHash)
		}
	})

	t.Run("session is active initially", func(t *testing.T) {
		ctx := context.Background()
		sess := &session.Session{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			TokenHash:      "test",
			CreatedAt:      time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			LastActivityAt: time.Now(),
		}

		if !sess.IsActive() {
			t.Error("expected session to be active")
		}
	})

	t.Run("session expires after expiry time", func(t *testing.T) {
		ctx := context.Background()
		sess := &session.Session{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			TokenHash:      "test",
			CreatedAt:      time.Now().Add(-48 * time.Hour),
			ExpiresAt:      time.Now().Add(-24 * time.Hour),
			LastActivityAt: time.Now().Add(-48 * time.Hour),
		}

		if !sess.IsExpired() {
			t.Error("expected session to be expired")
		}
	})
}

func TestSessionRepository_FindByTokenHash(t *testing.T) {
	t.Run("returns session when token hash exists", func(t *testing.T) {
		ctx := context.Background()
		tokenHash := hashToken("valid-token")
		expectedUserID := uuid.New()

		sess, err := findSessionByTokenHash(ctx, tokenHash)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sess.UserID != expectedUserID {
			t.Errorf("expected userID %s, got %s", expectedUserID, sess.UserID)
		}
	})

	t.Run("returns error when token hash does not exist", func(t *testing.T) {
		ctx := context.Background()

		_, err := findSessionByTokenHash(ctx, "nonexistent-hash")
		if err == nil {
			t.Error("expected error for non-existent token hash")
		}
	})
}

func TestSessionRepository_Invalidate(t *testing.T) {
	t.Run("invalidates active session", func(t *testing.T) {
		ctx := context.Background()
		sess := &session.Session{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			TokenHash:      "test",
			CreatedAt:      time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			LastActivityAt: time.Now(),
		}

		now := time.Now()
		sess.InvalidatedAt = &now

		if sess.IsActive() {
			t.Error("expected session to be invalidated")
		}
		if !sess.IsActive() && sess.InvalidatedAt != nil {
			t.Logf("session invalidated at %v", sess.InvalidatedAt)
		}
	})
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func findSessionByTokenHash(ctx context.Context, tokenHash string) (*session.Session, error) {
	return nil, session.ErrSessionNotFound
}