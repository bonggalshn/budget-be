package auth

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	t.Run("accepts password >= 8 characters", func(t *testing.T) {
		if !ValidatePassword("securepass123") {
			t.Error("expected password >= 8 chars to be valid")
		}
	})

	t.Run("rejects password < 8 characters", func(t *testing.T) {
		if ValidatePassword("short") {
			t.Error("expected password < 8 chars to be invalid")
		}
	})

	t.Run("rejects empty password", func(t *testing.T) {
		if ValidatePassword("") {
			t.Error("expected empty password to be invalid")
		}
	})

	t.Run("accepts password <= 72 characters", func(t *testing.T) {
		if !ValidatePassword("123456789012345678901234567890123456789012345678901234567890123456789012") {
			t.Error("expected password <= 72 chars to be valid")
		}
	})

	t.Run("rejects password > 72 characters", func(t *testing.T) {
		password := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123"
		if ValidatePassword(password) {
			t.Error("expected password > 72 chars to be invalid")
		}
	})
}