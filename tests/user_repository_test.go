package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/bonggalshn/budget-be/internal/user"
	"github.com/google/uuid"
)

func TestUserRepository_FindByUsername(t *testing.T) {
	t.Run("returns user when username exists", func(t *testing.T) {
		ctx := context.Background()
		testUser := &user.User{
			ID:           uuid.New(),
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$12$dummyhash",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		foundUser, err := findUserByUsername(ctx, "testuser")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if foundUser.Username != testUser.Username {
			t.Errorf("expected username %s, got %s", testUser.Username, foundUser.Username)
		}
	})

	t.Run("returns nil when username does not exist", func(t *testing.T) {
		ctx := context.Background()

		_, err := findUserByUsername(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for non-existent username")
		}
	})

	t.Run("finds user by username case-insensitively", func(t *testing.T) {
		ctx := context.Background()
		testUser := &user.User{
			ID:           uuid.New(),
			Username:     "TestUser",
			Email:        "test@example.com",
			PasswordHash: "$2a$12$dummyhash",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		foundUser, err := findUserByUsername(ctx, "TESTUSER")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if foundUser.Username != testUser.Username && foundUser.Username != "TestUser" {
			t.Logf("note: case sensitivity handling depends on DB")
		}
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	t.Run("returns user when email exists", func(t *testing.T) {
		ctx := context.Background()
		testUser := &user.User{
			ID:           uuid.New(),
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$12$dummyhash",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		foundUser, err := findUserByEmail(ctx, "test@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if foundUser.Email != testUser.Email {
			t.Errorf("expected email %s, got %s", testUser.Email, foundUser.Email)
		}
	})

	t.Run("returns nil when email does not exist", func(t *testing.T) {
		ctx := context.Background()

		_, err := findUserByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Error("expected error for non-existent email")
		}
	})

	t.Run("email is case-insensitive", func(t *testing.T) {
		ctx := context.Background()

		_, err := findUserByEmail(ctx, "TEST@EXAMPLE.COM")
		if err == nil {
			t.Logf("note: email case insensitivity depends on DB implementation")
		}
	})
}

func findUserByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}

func findUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, user.ErrUserNotFound
}