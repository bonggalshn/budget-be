-- Create verification_tokens table for email verification
CREATE TABLE IF NOT EXISTS verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for token lookup
CREATE INDEX IF NOT EXISTS idx_verification_token ON verification_tokens(token) WHERE expires_at > NOW();

-- Index for user cleanup
CREATE INDEX IF NOT EXISTS idx_verification_token_user ON verification_tokens(user_id);

COMMENT ON TABLE verification_tokens IS 'Stores email verification tokens for user registration';
COMMENT ON COLUMN verification_tokens.token IS 'Secure random token sent to user email';
COMMENT ON COLUMN verification_tokens.expires_at IS 'Token expiration time (24 hours from creation)';