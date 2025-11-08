
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(50) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    total_amount BIGINT NOT NULL CHECK (total_amount >= 0),
    discount_amount BIGINT DEFAULT 0 CHECK (discount_amount >= 0),
    shipping_fee BIGINT DEFAULT 0 CHECK (shipping_fee >= 0),
    pay_amount BIGINT NOT NULL CHECK (pay_amount >= 0),
    status VARCHAR(30) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'shipped', 'completed', 'cancelled', 'refunded')),
    payment_status VARCHAR(30) NOT NULL DEFAULT 'unpaid' CHECK (payment_status IN ('unpaid', 'paid', 'refunded')),
    ship_status VARCHAR(30) NOT NULL DEFAULT 'unshipped' CHECK (ship_status IN ('unshipped', 'shipped', 'received')),
    receiver_name VARCHAR(50) NOT NULL,
    receiver_phone VARCHAR(20) NOT NULL,
    receiver_address VARCHAR(500) NOT NULL,
    receiver_zip_code VARCHAR(20),
    remark TEXT,
    paid_at TIMESTAMPTZ,
    shipped_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

CREATE INDEX idx_orders_order_no ON orders(order_no);
CREATE INDEX idx_orders_user_id ON orders(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_status ON orders(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_payment_status ON orders(payment_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_ship_status ON orders(ship_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_created_at ON orders(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_deleted_at ON orders(deleted_at);


CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_name VARCHAR(200) NOT NULL,
    product_image VARCHAR(500),
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price BIGINT NOT NULL CHECK (unit_price >= 0),
    total_price BIGINT NOT NULL CHECK (total_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

CREATE INDEX idx_order_items_order_id ON order_items(order_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_order_items_product_id ON order_items(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_order_items_deleted_at ON order_items(deleted_at);
