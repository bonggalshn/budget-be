package session

import "errors"

var ErrSessionNotFound = errors.New("session not found")

type Repository interface {
	Create(ctx context.Context, s *Session) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	Invalidate(ctx context.Context, id string) error
	UpdateLastActivity(ctx context.Context, id string) error
}

type repository struct {
	db interface {
		Query(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
		Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
	}
}

func NewRepository(db interface{}) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, s *Session) error {
	return nil
}

func (r *repository) FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	return nil, ErrSessionNotFound
}

func (r *repository) Invalidate(ctx context.Context, id string) error {
	return nil
}

func (r *repository) UpdateLastActivity(ctx context.Context, id string) error {
	return nil
}