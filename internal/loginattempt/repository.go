package loginattempt

import (
	"context"
	"time"

	"github.com/bonggalshn/budget-be/internal/db"
	"github.com/google/uuid"
)

// Repository defines the interface for login attempt tracking.
type Repository interface {
	// Create records a new login attempt.
	Create(ctx context.Context, a *LoginAttempt) error
	// CountRecent returns the number of failed login attempts for a user since the given time.
	CountRecent(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	// CountRecentByIP returns the number of failed login attempts from an IP since the given time.
	CountRecentByIP(ctx context.Context, ipAddress string, since time.Time) (int, error)
}

type repository struct {
	db *db.Pool
}

func NewRepository(pool *db.Pool) Repository {
	return &repository{db: pool}
}

func (r *repository) Create(ctx context.Context, a *LoginAttempt) error {
	a.ID = uuid.New()
	a.AttemptedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`INSERT INTO login_attempts (id, user_id, identifier_provided, ip_address, success, attempted_at, failure_reason)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		a.ID, a.UserID, a.IdentifierProvided, a.IPAddress, a.Success, a.AttemptedAt, a.FailureReason,
	)
	return err
}

func (r *repository) CountRecent(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM login_attempts WHERE user_id = $1 AND success = false AND attempted_at > $2`,
		userID, since,
	).Scan(&count)
	return count, err
}

func (r *repository) CountRecentByIP(ctx context.Context, ipAddress string, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM login_attempts WHERE ip_address = $1 AND success = false AND attempted_at > $2`,
		ipAddress, since,
	).Scan(&count)
	return count, err
}
