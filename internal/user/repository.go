package user

import (
	"context"
	"errors"

	"github.com/bonggalshn/budget-be/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrUserNotFound is returned when a user is not found in the database.
var ErrUserNotFound = errors.New("user not found")

// Repository defines the interface for user data access operations.
type Repository interface {
	// Create inserts a new user into the database.
	Create(ctx context.Context, u *User) error

	// CreateVerificationToken creates a new email verification token.
	CreateVerificationToken(ctx context.Context, vt *VerificationToken) error

	// MarkEmailVerified marks a user's email as verified.
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error

	// FindByToken finds a user by their verification token.
	FindByToken(ctx context.Context, token string) (*VerificationToken, *User, error)

	// FindByUsername retrieves a user by their username.
	// Returns ErrUserNotFound if no user exists with the given username.
	FindByUsername(ctx context.Context, username string) (*User, error)

	// FindByEmail retrieves a user by their email address.
	// Returns ErrUserNotFound if no user exists with the given email.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// FindByID retrieves a user by their unique identifier.
	// Returns ErrUserNotFound if no user exists with the given ID.
	FindByID(ctx context.Context, id string) (*User, error)
}

type repository struct {
	db *db.Pool
}

func NewRepository(pool *db.Pool) Repository {
	return &repository{db: pool}
}

func (r *repository) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, u.ID, u.Username, u.Email, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *repository) CreateVerificationToken(ctx context.Context, vt *VerificationToken) error {
	query := `
		INSERT INTO verification_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, vt.ID, vt.UserID, vt.Token, vt.ExpiresAt, vt.CreatedAt)
	return err
}

func (r *repository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *repository) FindByToken(ctx context.Context, token string) (*VerificationToken, *User, error) {
	query := `
		SELECT vt.id, vt.user_id, vt.token, vt.expires_at, vt.created_at,
		       u.id, u.username, u.email, u.password_hash, u.email_verified, u.created_at, u.updated_at, u.deleted_at
		FROM verification_tokens vt
		JOIN users u ON vt.user_id = u.id
		WHERE vt.token = $1 AND vt.expires_at > NOW() AND u.deleted_at IS NULL
	`
	row := r.db.QueryRow(ctx, query, token)
	var vt VerificationToken
	var u User
	err := row.Scan(
		&vt.ID, &vt.UserID, &vt.Token, &vt.ExpiresAt, &vt.CreatedAt,
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrUserNotFound
		}
		return nil, nil, err
	}
	return &vt, &u, nil
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
