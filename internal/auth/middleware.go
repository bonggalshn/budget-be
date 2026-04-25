package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bonggalshn/budget-be/internal/config"
)

type Middleware struct {
	service *Service
	config  config.Config
}

func NewMiddleware(service *Service, cfg config.Config) *Middleware {
	return &Middleware{
		service: service,
		config:  cfg,
	}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getToken(r)
		if token == "" {
			m.writeError(w, "unauthorized", "Missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := m.service.ValidateToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, ErrSessionExpired) {
				m.writeError(w, "session_expired", "Session expired. Please log in again.", http.StatusUnauthorized)
				return
			}
			m.writeError(w, "unauthorized", "Missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) writeError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(middlewareError{
		ErrorCode: code,
		Message: message,
	})
}

type middlewareError struct {
	ErrorCode string `json:"error_code"`
	Message  string `json:"message"`
	Details  any    `json:"details,omitempty"`
}