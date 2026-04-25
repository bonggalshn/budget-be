package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bonggalshn/budget-be/internal/config"
)

type Handler struct {
	service *Service
	config  config.Config
}

func NewHandler(service *Service, cfg config.Config) *Handler {
	return &Handler{
		service: service,
		config:  cfg,
	}
}

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
