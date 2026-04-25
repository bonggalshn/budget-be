# Run Registration Migrations
# Usage: Run this in SQL Server Management Studio, psql, or any PostgreSQL client

# ============================================
# Migration 004: Add Email Verification Fields
# ============================================

ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_user_unverified ON users(email_verified) WHERE email_verified = FALSE;

COMMENT ON COLUMN users.email_verified IS 'Whether the user has verified their email address';

-- ============================================
# Migration 005: Create Verification Tokens Table
# ============================================

CREATE TABLE IF NOT EXISTS verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verification_token ON verification_tokens(token) WHERE expires_at > NOW();

CREATE INDEX IF NOT EXISTS idx_verification_token_user ON verification_tokens(user_id);

COMMENT ON TABLE verification_tokens IS 'Stores email verification tokens for user registration';
COMMENT ON COLUMN verification_tokens.token IS 'Secure random token sent to user email';
COMMENT ON COLUMN verification_tokens.expires_at IS 'Token expiration time (24 hours from creation)';