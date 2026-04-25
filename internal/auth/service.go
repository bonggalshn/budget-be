package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid username/email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionExpired     = errors.New("session expired. please log in again")
	ErrAccountLocked     = errors.New("account temporarily locked. try again in 15 minutes")
)

type Config struct {
	Secret          string
	Expiry          time.Duration
}

type Service struct {
	config           Config
	userRepo        UserRepository
	sessionRepo    SessionRepository
	attemptRepo    LoginAttemptRepository
}

type UserRepository interface {
	FindByUsername(ctx interface{}, username string) (*User, error)
	FindByEmail(ctx interface{}, email string) (*User, error)
}

type SessionRepository interface {
	Create(ctx interface{}, s *Session) error
	FindByTokenHash(ctx interface{}, tokenHash string) (*Session, error)
	Invalidate(ctx interface{}, id string) error
}

type LoginAttemptRepository interface {
	Create(ctx interface{}, a *LoginAttempt) error
	CountRecent(ctx interface{}, userID string, window time.Duration) (int, error)
	CountByIP(ctx interface{}, ip string, window time.Duration) (int, error)
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password  string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

type UserInfo struct {
	ID        string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func IsValidPassword(password string) bool {
	return len(password) >= 8
}