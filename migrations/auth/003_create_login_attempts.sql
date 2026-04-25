-- Create login_attempts table for audit trail and rate limiting
CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    identifier_provided VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    success BOOLEAN NOT NULL,
    attempted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    failure_reason VARCHAR(50)
);

-- Indexes for rate limiting queries
CREATE INDEX IF NOT EXISTS idx_login_attempt_user_id ON login_attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_login_attempt_attempted_at ON login_attempts(attempted_at);
CREATE INDEX IF NOT EXISTS idx_login_attempt_ip_ident ON login_attempts(ip_address, identifier_provided, attempted_at);