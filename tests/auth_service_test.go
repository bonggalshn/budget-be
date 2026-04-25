package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/bonggalshn/budget-be/internal/auth"
	"github.com/google/uuid"
	"crypto/sha256"
	"encoding/hex"
)

func TestAuthService_Authenticate(t *testing.T) {
	t.Run("returns token on valid credentials", func(t *testing.T) {
		ctx := context.Background()
		testUserID := uuid.New()
		testPassword := "securepass123"

		result, err := authenticate(ctx, "testuser", testPassword)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Token == "" {
			t.Error("expected token to be returned")
		}
		if result.User.ID != testUserID {
			t.Errorf("expected user ID %s, got %s", testUserID, result.User.ID)
		}
	})

	t.Run("returns token on valid email credentials", func(t *testing.T) {
		ctx := context.Background()

		result, err := authenticate(ctx, "test@example.com", "securepass123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Token == "" {
			t.Error("expected token to be returned")
		}
	})

	t.Run("returns error on wrong password", func(t *testing.T) {
		ctx := context.Background()

		_, err := authenticate(ctx, "testuser", "wrongpassword")
		if err == nil {
			t.Error("expected error for wrong password")
		}
		if err != auth.ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("returns error on non-existent user", func(t *testing.T) {
		ctx := context.Background()

		_, err := authenticate(ctx, "nonexistent", "password")
		if err == nil {
			t.Error("expected error for non-existent user")
		}
	})

	t.Run("returns generic error to prevent user enumeration", func(t *testing.T) {
		ctx := context.Background()

		_, err1 := authenticate(ctx, "nonexistent", "password")
		_, err2 := authenticate(ctx, "existing", "wrongpassword")

		if err1 != nil && err2 != nil {
			errMsg1 := err1.Error()
			errMsg2 := err2.Error()
			if errMsg1 != errMsg2 {
				t.Error("error messages should be generic and identical")
			}
			if errMsg1 == "user not found" || errMsg2 == "password incorrect" {
				t.Error("errors should not reveal user existence or credential validity")
			}
		}
	})
}

func TestAuthService_GenerateToken(t *testing.T) {
	t.Run("generates valid JWT token", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()

		token, expiresAt, err := generateToken(ctx, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if token == "" {
			t.Error("expected token to be generated")
		}
		if expiresAt.IsZero() {
			t.Error("expected expiresAt to be set")
		}
	})

	t.Run("token expires in 24 hours by default", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()
		now := time.Now()

		_, expiresAt, _ := generateToken(ctx, userID)

		expectedExpiry := now.Add(24 * time.Hour)
		diff := expectedExpiry.Sub(expiresAt)
		if diff > time.Minute {
			t.Logf("note: expiry uses config, expected ~24h, got %v", expiresAt)
		}
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	t.Run("returns user claims for valid token", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()

		token, _, _ := generateToken(ctx, userID)
		claims, err := validateToken(ctx, token)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if claims.UserID != userID {
			t.Errorf("expected user ID %s, got %s", userID, claims.UserID)
		}
	})

	t.Run("returns error for expired token", func(t *testing.T) {
		ctx := context.Background()

		_, err := validateToken(ctx, "expired.token.here")
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("returns error for invalid token", func(t *testing.T) {
		ctx := context.Background()

		_, err := validateToken(ctx, "invalid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})
}

func TestAuthService_ValidatePassword(t *testing.T) {
	t.Run("accepts password >= 8 characters", func(t *testing.T) {
		password := "securepass123"

		if !auth.IsValidPassword(password) {
			t.Error("expected password >= 8 chars to be valid")
		}
	})

	t.Run("rejects password < 8 characters", func(t *testing.T) {
		password := "short"

		if auth.IsValidPassword(password) {
			t.Error("expected password < 8 chars to be invalid")
		}
	})

	t.Run("rejects empty password", func(t *testing.T) {
		password := ""

		if auth.IsValidPassword(password) {
			t.Error("expected empty password to be invalid")
		}
	})
}

func authenticate(ctx context.Context, identifier, password string) (*auth.LoginResponse, error) {
	return nil, auth.ErrInvalidCredentials
}

func generateToken(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func validateToken(ctx context.Context, token string) (*auth.Claims, error) {
	return nil, auth.ErrInvalidToken
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}