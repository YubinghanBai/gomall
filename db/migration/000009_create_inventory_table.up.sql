-- Inventory table: tracks product stock levels and movements
CREATE TABLE IF NOT EXISTS inventory (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL UNIQUE REFERENCES products(id) ON DELETE RESTRICT,
    available_stock INT NOT NULL DEFAULT 0 CHECK (available_stock >= 0),
    reserved_stock INT NOT NULL DEFAULT 0 CHECK (reserved_stock >= 0),
    total_stock INT NOT NULL GENERATED ALWAYS AS (available_stock + reserved_stock) STORED,
    low_stock_threshold INT DEFAULT 10,
    version BIGINT NOT NULL DEFAULT 0, -- For optimistic locking
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT inventory_stock_non_negative CHECK (available_stock >= 0 AND reserved_stock >= 0)
);

-- Inventory logs table: records all stock movements for auditing
CREATE TABLE IF NOT EXISTS inventory_logs (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    change_type VARCHAR(20) NOT NULL CHECK (change_type IN ('restock', 'reserve', 'release', 'deduct', 'adjust')),
    quantity_change INT NOT NULL,
    before_available INT NOT NULL,
    after_available INT NOT NULL,
    before_reserved INT NOT NULL,
    after_reserved INT NOT NULL,
    reason VARCHAR(500),
    operator_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Inventory reservations table: tracks temporary stock reservations (e.g., for pending orders)
CREATE TABLE IF NOT EXISTS inventory_reservations (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK (quantity > 0),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'confirmed', 'cancelled', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(product_id, order_id)
);

-- Indexes for inventory
CREATE INDEX idx_inventory_product_id ON inventory(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inventory_available_stock ON inventory(available_stock) WHERE deleted_at IS NULL;
CREATE INDEX idx_inventory_deleted_at ON inventory(deleted_at);

-- Indexes for inventory_logs
CREATE INDEX idx_inventory_logs_product_id ON inventory_logs(product_id);
CREATE INDEX idx_inventory_logs_order_id ON inventory_logs(order_id) WHERE order_id IS NOT NULL;
CREATE INDEX idx_inventory_logs_created_at ON inventory_logs(created_at DESC);
CREATE INDEX idx_inventory_logs_change_type ON inventory_logs(change_type);

-- Indexes for inventory_reservations
CREATE INDEX idx_inventory_reservations_product_id ON inventory_reservations(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inventory_reservations_order_id ON inventory_reservations(order_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_inventory_reservations_status ON inventory_reservations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_inventory_reservations_expires_at ON inventory_reservations(expires_at) WHERE status = 'active' AND deleted_at IS NULL;
CREATE INDEX idx_inventory_reservations_deleted_at ON inventory_reservations(deleted_at);

-- Function to update inventory updated_at timestamp
CREATE OR REPLACE FUNCTION update_inventory_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for inventory table
CREATE TRIGGER trigger_update_inventory_updated_at
    BEFORE UPDATE ON inventory
    FOR EACH ROW
    EXECUTE FUNCTION update_inventory_updated_at();

-- Trigger for inventory_reservations table
CREATE TRIGGER trigger_update_inventory_reservations_updated_at
    BEFORE UPDATE ON inventory_reservations
    FOR EACH ROW
    EXECUTE FUNCTION update_inventory_updated_at();
