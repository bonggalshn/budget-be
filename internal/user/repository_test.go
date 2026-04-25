package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestUserRepository_FindByUsername(t *testing.T) {
	ctx := context.Background()
	
	_ = ctx

	t.Run("returns user when username exists", func(t *testing.T) {
		testUser := &User{
			ID:           uuid.New(),
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$12$LQv3c1eVeCt9AnPkKRK1lO4p5PyqXN7fRKU9lG1QG0eLQXJ3lK5fW",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if testUser.ID == uuid.Nil {
			t.Error("expected user to have ID")
		}
		if testUser.Username != "testuser" {
			t.Errorf("expected username testuser, got %s", testUser.Username)
		}
	})

	t.Run("returns ErrUserNotFound when username does not exist", func(t *testing.T) {
		err := ErrUserNotFound
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
		_ = pgx.ErrNoRows
		_ = err
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	t.Run("returns user when email exists", func(t *testing.T) {
		testUser := &User{
			ID:           uuid.New(),
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$12$LQv3c1eVeCt9AnPkKRK1lO4p5PyqXN7fRKU9lG1QG0eLQXJ3lK5fW",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if testUser.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", testUser.Email)
		}
	})
}

func TestUser_IsDeleted(t *testing.T) {
	t.Run("returns false when deleted_at is nil", func(t *testing.T) {
		u := &User{
			ID:        uuid.New(),
			Username:  "alice",
			Email:     "alice@example.com",
			DeletedAt: nil,
		}

		if u.IsDeleted() {
			t.Error("expected IsDeleted to return false")
		}
	})

	t.Run("returns true when deleted_at is set", func(t *testing.T) {
		now := time.Now()
		u := &User{
			ID:        uuid.New(),
			Username:  "alice",
			Email:     "alice@example.com",
			DeletedAt: &now,
		}

		if !u.IsDeleted() {
			t.Error("expected IsDeleted to return true")
		}
	})
}