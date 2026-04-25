-- Add email verification fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Add index for unverified users lookup
CREATE INDEX IF NOT EXISTS idx_user_unverified ON users(email_verified) WHERE email_verified = FALSE;

COMMENT ON COLUMN users.email_verified IS 'Whether the user has verified their email address';