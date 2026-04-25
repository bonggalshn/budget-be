package session

import (
	"context"
	"errors"
	"time"

	"github.com/bonggalshn/budget-be/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrSessionNotFound = errors.New("session not found")

type Repository interface {
	Create(ctx context.Context, s *Session) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	Invalidate(ctx context.Context, id string) error
	UpdateLastActivity(ctx context.Context, id string) error
}

type repository struct {
	db *db.Pool
}

func NewRepository(pool *db.Pool) Repository {
	return &repository{db: pool}
}

func (r *repository) Create(ctx context.Context, s *Session) error {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.LastActivityAt = s.CreatedAt

	_, err := r.db.Exec(ctx,
		`INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at, invalidated_at, last_activity_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		s.ID, s.UserID, s.TokenHash, s.CreatedAt, s.ExpiresAt, s.InvalidatedAt, s.LastActivityAt,
	)
	return err
}

func (r *repository) FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, created_at, expires_at, invalidated_at, last_activity_at
		 FROM sessions WHERE token_hash = $1`,
		tokenHash,
	)
	var s Session
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.TokenHash,
		&s.CreatedAt,
		&s.ExpiresAt,
		&s.InvalidatedAt,
		&s.LastActivityAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) Invalidate(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx,
		`UPDATE sessions SET invalidated_at = $1 WHERE id = $2`,
		time.Now(), parsedID,
	)
	return err
}

func (r *repository) UpdateLastActivity(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx,
		`UPDATE sessions SET last_activity_at = $1 WHERE id = $2`,
		time.Now(), parsedID,
	)
	return err
}