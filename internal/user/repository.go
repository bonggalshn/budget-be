package user

import (
	"context"
	"errors"

	"github.com/bonggalshn/budget-be/internal/db"
	"github.com/jackc/pgx/v5"
)

var ErrUserNotFound = errors.New("user not found")

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type repository struct {
	db *db.Pool
}

func NewRepository(pool *db.Pool) Repository {
	return &repository{db: pool}
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	u, err := r.findUser(ctx, "SELECT id, username, email, password_hash, created_at, updated_at, deleted_at FROM users WHERE username = $1 AND deleted_at IS NULL", username)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	u, err := r.findUser(ctx, "SELECT id, username, email, password_hash, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS NULL", email)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*User, error) {
	u, err := r.findUser(ctx, "SELECT id, username, email, password_hash, created_at, updated_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *repository) findUser(ctx context.Context, query string, args ...interface{}) (*User, error) {
	row := r.db.QueryRow(ctx, query, args...)
	var u User
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}