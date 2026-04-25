package user

import "errors"

var ErrUserNotFound = errors.New("user not found")

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type repository struct {
	db interface {
		Query(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
	}
}

func NewRepository(db interface{}) Repository {
	return &repository{db: db}
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	return nil, ErrUserNotFound
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	return nil, ErrUserNotFound
}

func (r *repository) FindByID(ctx context.Context, id string) (*User, error) {
	return nil, ErrUserNotFound
}