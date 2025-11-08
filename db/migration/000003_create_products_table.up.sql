CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    brand VARCHAR(100),
    price BIGINT NOT NULL CHECK (price >= 0),
    origin_price BIGINT NOT NULL CHECK (origin_price >= 0),
    cost_price BIGINT CHECK (cost_price >= 0),
    stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    low_stock_threshold INT DEFAULT 10,
    sales_count INT DEFAULT 0,
    view_count INT DEFAULT 0,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'off_shelf')),
    is_featured BOOLEAN DEFAULT FALSE,
    specifications JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

CREATE INDEX idx_products_name ON products(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_brand ON products(brand) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category_id ON products(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_status ON products(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_is_featured ON products(is_featured) WHERE deleted_at IS NULL AND is_featured = TRUE;
CREATE INDEX idx_products_sales_count ON products(sales_count DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_deleted_at ON products(deleted_at);

CREATE TABLE IF NOT EXISTS product_images (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url VARCHAR(500) NOT NULL,
    sort INT DEFAULT 0,
    is_main BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

CREATE INDEX idx_product_images_product_id ON product_images(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_images_deleted_at ON product_images(deleted_at);

