CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE,
    parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    icon VARCHAR(500),
    sort INT DEFAULT 0,
    level SMALLINT DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

CREATE INDEX idx_categories_parent_id ON categories(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_level ON categories(level);
CREATE INDEX idx_categories_is_active ON categories(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_deleted_at ON categories(deleted_at);