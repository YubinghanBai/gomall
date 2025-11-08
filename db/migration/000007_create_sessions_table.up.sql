CREATE TABLE IF NOT EXISTS sessions (
                                        id UUID PRIMARY KEY,
                                        user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) NOT NULL,
    user_agent VARCHAR(500) NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);