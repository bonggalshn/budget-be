package loginattempt

import (
	"time"

	"github.com/google/uuid"
)

type LoginAttempt struct {
	ID               uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id,omitempty"`
	IdentifierProvided string `json:"identifier_provided"`
	IPAddress       string    `json:"ip_address"`
	Success         bool      `json:"success"`
	AttemptedAt     time.Time `json:"attempted_at"`
	FailureReason   *string  `json:"failure_reason,omitempty"`
}

const (
	FailureInvalidCredentials = "invalid_credentials"
	FailureAccountLocked     = "account_locked"
)