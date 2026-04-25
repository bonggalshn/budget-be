package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/bonggalshn/budget-be/internal/config"
)

// Handler handles HTTP requests for authentication endpoints.
type Handler struct {
	service *Service
	config  config.Config
}

// NewHandler creates a new authentication handler with the given service and configuration.
func NewHandler(service *Service, cfg config.Config) *Handler {
	return &Handler{
		service: service,
		config:  cfg,
	}
}

// Login handles POST /api/v1/auth/login requests.
// Validates credentials and returns a JWT token on success.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Identifier == "" || req.Password == "" {
		h.writeError(w, "invalid_request", "Missing required field: identifier or password", http.StatusBadRequest)
		return
	}

	ipAddress := getClientIP(r)
	resp, err := h.service.Authenticate(r.Context(), req.Identifier, req.Password, ipAddress)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			h.writeError(w, "invalid_credentials", "Invalid username/email or password", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, ErrAccountLocked) {
			h.writeError(w, "account_locked", "Account temporarily locked. Try again in 15 minutes.", http.StatusTooManyRequests)
			return
		}
		h.writeError(w, "service_unavailable", "Service temporarily unavailable. Please try again.", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Register handles POST /api/v1/auth/register requests.
// Creates a new user account.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		h.writeError(w, "invalid_request", "Missing required field: username, email, or password", http.StatusBadRequest)
		return
	}

	resp, err := h.service.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		log.Printf("Register error: %v", err)
		if errors.Is(err, ErrEmailAlreadyExists) {
			h.writeError(w, "duplicate_email", "Email already registered", http.StatusConflict)
			return
		}
		if errors.Is(err, ErrUsernameTaken) {
			h.writeError(w, "username_taken", "Username already taken", http.StatusConflict)
			return
		}
		if errors.Is(err, ErrInvalidEmail) {
			h.writeError(w, "invalid_email", "Invalid email format", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrWeakPassword) {
			h.writeError(w, "weak_password", "Password must be at least 8 characters with at least one number", http.StatusBadRequest)
			return
		}
		h.writeError(w, "service_unavailable", "Service temporarily unavailable. Please try again.", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Verify handles POST /api/v1/auth/verify requests.
// Verifies a user's email address.
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		h.writeError(w, "invalid_request", "Missing required field: token", http.StatusBadRequest)
		return
	}

	if err := h.service.VerifyEmail(r.Context(), req.Token); err != nil {
		if errors.Is(err, ErrInvalidVerificationToken) {
			h.writeError(w, "invalid_token", "Verification token invalid or expired", http.StatusBadRequest)
			return
		}
		h.writeError(w, "service_unavailable", "Service temporarily unavailable. Please try again.", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VerifyResponse{
		Message: "Email verified successfully",
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := getToken(r)
	if token == "" {
		h.writeError(w, "unauthorized", "Missing or invalid authorization header", http.StatusUnauthorized)
		return
	}

	if err := h.service.Logout(r.Context(), token); err != nil {
		h.writeError(w, "service_unavailable", "Service temporarily unavailable. Please try again.", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		h.writeError(w, "unauthorized", "Missing or invalid authorization header", http.StatusUnauthorized)
		return
	}

	u, err := h.service.UserRepo.FindByID(r.Context(), userID)
	if err != nil || u == nil {
		h.writeError(w, "unauthorized", "User not found", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserInfo{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	})
}

func (h *Handler) writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(authError{
		ErrorCode: code,
		Message:   message,
	})
}

type authError struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
}
