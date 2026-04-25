package auth

import (
	"context"
	"crypto/rand"
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

// Error definitions for authentication failures.
var (
	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid username/email or password")

	// ErrInvalidToken is returned when a JWT token is invalid or malformed.
	ErrInvalidToken = errors.New("invalid or expired token")

	// ErrSessionExpired is returned when a session has exceeded its expiry time.
	ErrSessionExpired = errors.New("session expired. please log in again")

	// ErrAccountLocked is returned when an account has been locked due to too many failed attempts.
	ErrAccountLocked = errors.New("account temporarily locked. try again in 15 minutes")

	// ErrNotFound is returned when a user is not found.
	ErrNotFound = errors.New("user not found")

	// ErrEmailAlreadyExists is returned when email is already registered.
	ErrEmailAlreadyExists = errors.New("email already registered")

	// ErrUsernameTaken is returned when username is already taken.
	ErrUsernameTaken = errors.New("username already taken")

	// ErrInvalidEmail is returned when email format is invalid.
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrWeakPassword is returned when password does not meet requirements.
	ErrWeakPassword = errors.New("password must be at least 8 characters with at least one number")

	// ErrInvalidVerificationToken is returned when verification token is invalid or expired.
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")

	// ErrEmailNotVerified is returned when user tries to login without verifying email.
	ErrEmailNotVerified = errors.New("email not verified")
)

// Config holds authentication service configuration settings.
type Config struct {
	// Secret is the JWT signing key.
	Secret string
	// Expiry is the token expiration duration.
	Expiry time.Duration
}

// Service handles user authentication and session management.
type Service struct {
	cfg         config.Config
	UserRepo    user.Repository
	sessionRepo session.Repository
	attemptRepo loginattempt.Repository
}

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	CreateVerificationToken(ctx context.Context, vt *user.VerificationToken) error
	FindByUsername(ctx context.Context, username string) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByID(ctx context.Context, id string) (*user.User, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	FindByToken(ctx context.Context, token string) (*user.VerificationToken, *user.User, error)
}

// SessionRepository defines the interface for session data access.
type SessionRepository interface {
	Create(ctx context.Context, s *session.Session) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*session.Session, error)
	Invalidate(ctx context.Context, id string) error
	UpdateLastActivity(ctx context.Context, id string) error
}

// LoginAttemptRepository defines the interface for login attempt tracking.
type LoginAttemptRepository interface {
	Create(ctx context.Context, a *loginattempt.LoginAttempt) error
	CountRecent(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	CountRecentByIP(ctx context.Context, ipAddress string, since time.Time) (int, error)
}

// LoginRequest represents a login API request payload.
type LoginRequest struct {
	Identifier string `json:"identifier"` // Username or email address
	Password   string `json:"password"`   // User's password
}

// RegisterRequest represents a user registration API request payload.
type RegisterRequest struct {
	Username string `json:"username"` // Desired username
	Email    string `json:"email"`    // Email address
	Password string `json:"password"` // User's password
}

// VerifyRequest represents an email verification API request payload.
type VerifyRequest struct {
	Token string `json:"token"` // Verification token from email
}

// VerifyResponse represents a successful email verification API response.
type VerifyResponse struct {
	Message string `json:"message"` // Success message
}

// RegisterResponse represents a successful registration API response.
type RegisterResponse struct {
	Message string `json:"message"` // Success message
	UserID  string `json:"user_id"` // Created user ID
}

// LoginResponse represents a successful login API response.
type LoginResponse struct {
	Token     string    `json:"token"`      // JWT authentication token
	ExpiresAt time.Time `json:"expires_at"` // Token expiration timestamp
	User      UserInfo  `json:"user"`       // Authenticated user details
}

// UserInfo represents user information exposed in API responses.
type UserInfo struct {
	ID        string    `json:"id"`         // User's unique identifier
	Username  string    `json:"username"`   // Username
	Email     string    `json:"email"`      // Email address
	CreatedAt time.Time `json:"created_at"` // Account creation timestamp
}

// Claims represents JWT token claims for authenticated users.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewService(cfg config.Config, userRepo user.Repository, sessionRepo session.Repository, attemptRepo loginattempt.Repository) *Service {
	return &Service{
		cfg:         cfg,
		UserRepo:    userRepo,
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

	if !u.EmailVerified {
		return nil, ErrEmailNotVerified
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
			CreatedAt: u.CreatedAt,
		},
	}, nil
}

func (s *Service) Register(ctx context.Context, username, email, password string) (*RegisterResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := validateEmail(email); err != nil {
		return nil, ErrInvalidEmail
	}

	if err := validatePassword(password); err != nil {
		return nil, ErrWeakPassword
	}

	if _, err := s.UserRepo.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}

	if _, err := s.UserRepo.FindByUsername(ctx, username); err == nil {
		return nil, ErrUsernameTaken
	} else if !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	u := &user.User{
		ID:            uuid.New(),
		Username:      username,
		Email:         email,
		PasswordHash:  string(hash),
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.UserRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		Message: "Registration successful. Please verify your email.",
		UserID:  u.ID.String(),
	}, nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email required")
	}
	if len(email) < 3 || len(email) > 254 {
		return errors.New("email length invalid")
	}
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	if atIndex < 1 || atIndex == len(email)-1 {
		return errors.New("invalid email format")
	}
	domain := email[atIndex+1:]
	dotIndex := -1
	for i, c := range domain {
		if c == '.' {
			dotIndex = i
			break
		}
	}
	if dotIndex < 1 || dotIndex == len(domain)-1 {
		return errors.New("invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password too short")
	}
	hasNumber := false
	for _, c := range password {
		if c >= '0' && c <= '9' {
			hasNumber = true
			break
		}
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	return nil
}

func generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Service) CreateVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := generateVerificationToken()
	if err != nil {
		return "", err
	}

	vt := &user.VerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.UserRepo.CreateVerificationToken(ctx, vt); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) VerifyEmail(ctx context.Context, tokenStr string) error {
	vt, u, err := s.UserRepo.FindByToken(ctx, tokenStr)
	if err != nil {
		return ErrInvalidVerificationToken
	}

	if vt.IsExpired() {
		return ErrInvalidVerificationToken
	}

	if err := s.UserRepo.MarkEmailVerified(ctx, u.ID); err != nil {
		return err
	}

	return nil
}

func (s *Service) GenerateToken(ctx context.Context, userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.JWT.Expiry)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
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
		UserID:             userID,
		IdentifierProvided: identifier,
		IPAddress:          ipAddress,
		Success:            false,
		FailureReason:      strp("invalid_credentials"),
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

type TestUser struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type TestSession struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	TokenHash      string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	InvalidatedAt  *time.Time
	LastActivityAt time.Time
}

type TestAttempt struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	IdentifierProvided string
	IPAddress          string
	Success            bool
	AttemptedAt        time.Time
	FailureReason      *string
}

func NewServiceWithMocks(cfg config.Config, userRepo UserRepository, sessionRepo SessionRepository, attemptRepo LoginAttemptRepository) *Service {
	return &Service{
		cfg:         cfg,
		UserRepo:    userRepo,
		sessionRepo: sessionRepo,
		attemptRepo: attemptRepo,
	}
}

func (s *Service) TestAuthenticate(ctx context.Context, identifier, password, ipAddress string) (*LoginResponse, error) {
	return s.Authenticate(ctx, identifier, password, ipAddress)
}

func ValidatePassword(password string) bool {
	return len(password) >= 8 && len(password) <= 72
}
