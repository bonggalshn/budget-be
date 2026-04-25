package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/bonggalshn/budget-be/internal/config"
	"github.com/bonggalshn/budget-be/internal/loginattempt"
	"github.com/bonggalshn/budget-be/internal/session"
	"github.com/bonggalshn/budget-be/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username/email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionExpired     = errors.New("session expired. please log in again")
	ErrAccountLocked     = errors.New("account temporarily locked. try again in 15 minutes")
	ErrNotFound          = errors.New("user not found")
)

type Config struct {
	Secret          string
	Expiry          time.Duration
}

type Service struct {
	cfg         config.Config
	UserRepo   user.Repository
	sessionRepo session.Repository
	attemptRepo loginattempt.Repository
}

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByID(ctx context.Context, id string) (*user.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, s *session.Session) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*session.Session, error)
	Invalidate(ctx context.Context, id string) error
}

type LoginAttemptRepository interface {
	Create(ctx context.Context, a *loginattempt.LoginAttempt) error
	CountRecent(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	CountRecentByIP(ctx context.Context, ipAddress string, since time.Time) (int, error)
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
	ID        string    `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewService(cfg config.Config, userRepo user.Repository, sessionRepo session.Repository, attemptRepo loginattempt.Repository) *Service {
	return &Service{
		cfg:         cfg,
		UserRepo:   userRepo,
		sessionRepo: sessionRepo,
		attemptRepo: attemptRepo,
	}
}

func (s *Service) Authenticate(ctx context.Context, identifier, password, ipAddress string) (*LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var u *user.User
	var err error

	if u, err = s.UserRepo.FindByUsername(ctx, identifier); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			u, err = s.UserRepo.FindByEmail(ctx, identifier)
		}
	}
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	lockoutWindow := time.Now().Add(-15 * time.Minute)
	failedAttempts, _ := s.attemptRepo.CountRecent(ctx, u.ID, lockoutWindow)
	if failedAttempts >= 5 {
		return nil, ErrAccountLocked
	}

	if err := s.comparePassword(u.PasswordHash, password); err != nil {
		s.recordFailedAttempt(ctx, u.ID, identifier, ipAddress)
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := s.GenerateToken(ctx, u.ID.String())
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:        u.ID.String(),
			Username:  u.Username,
			Email:     u.Email,
			CreatedAt:  u.CreatedAt,
		},
	}, nil
}

func (s *Service) GenerateToken(ctx context.Context, userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.JWT.Expiry)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	tokenHash := hashToken(tokenString)
	session := &session.Session{
		UserID:    uuid.MustParse(userID),
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	tokenHash := hashToken(tokenString)
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !session.IsActive() {
		return nil, ErrSessionExpired
	}

	return claims, nil
}

func (s *Service) Logout(ctx context.Context, tokenString string) error {
	tokenHash := hashToken(tokenString)
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	return s.sessionRepo.Invalidate(ctx, session.ID.String())
}

func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *Service) comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *Service) recordFailedAttempt(ctx context.Context, userID uuid.UUID, identifier, ipAddress string) {
	attempt := &loginattempt.LoginAttempt{
		UserID:           userID,
		IdentifierProvided: identifier,
		IPAddress:        ipAddress,
		Success:         false,
		FailureReason:   strp("invalid_credentials"),
	}
	s.attemptRepo.Create(ctx, attempt)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func strp(s string) *string {
	return &s
}