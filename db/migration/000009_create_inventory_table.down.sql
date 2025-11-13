-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_inventory_reservations_updated_at ON inventory_reservations;
DROP TRIGGER IF EXISTS trigger_update_inventory_updated_at ON inventory;

-- Drop function
DROP FUNCTION IF EXISTS update_inventory_updated_at();

-- Drop tables
DROP TABLE IF EXISTS inventory_reservations;
DROP TABLE IF EXISTS inventory_logs;
DROP TABLE IF EXISTS inventory;
