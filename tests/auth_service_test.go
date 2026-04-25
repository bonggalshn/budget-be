package tests

import (
	"testing"

	"github.com/bonggalshn/budget-be/internal/auth"
)

func TestAuthService_ValidatePassword(t *testing.T) {
	t.Run("accepts password >= 8 characters", func(t *testing.T) {
		if !auth.ValidatePassword("securepass123") {
			t.Error("expected password >= 8 chars to be valid")
		}
	})

	t.Run("rejects password < 8 characters", func(t *testing.T) {
		if auth.ValidatePassword("short") {
			t.Error("expected password < 8 chars to be invalid")
		}
	})

	t.Run("rejects empty password", func(t *testing.T) {
		if auth.ValidatePassword("") {
			t.Error("expected empty password to be invalid")
		}
	})
}