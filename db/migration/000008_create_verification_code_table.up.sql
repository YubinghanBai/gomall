
CREATE TABLE IF NOT EXISTS verification_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(6) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'email_verification', 'password_reset'
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX idx_verification_codes_email ON verification_codes(email);
CREATE INDEX idx_verification_codes_code ON verification_codes(code);
CREATE INDEX idx_verification_codes_type ON verification_codes(type);