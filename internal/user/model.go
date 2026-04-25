package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the system.
type User struct {
	ID           uuid.UUID  `json:"id"`           // Unique identifier
	Username     string     `json:"username"`   // Login username (unique)
	Email        string     `json:"email"`        // Email address (unique)
	PasswordHash string     `json:"-"`          // Bcrypt hashed password (never exposed)
	CreatedAt    time.Time  `json:"created_at"` // Account creation timestamp
	UpdatedAt    time.Time  `json:"updated_at"` // Last update timestamp
	DeletedAt    *time.Time `json:"deleted_at,omitempty"` // Soft delete marker
}

// IsDeleted returns true if the user has been soft-deleted.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}
