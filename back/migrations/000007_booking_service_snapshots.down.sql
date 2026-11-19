DROP INDEX IF EXISTS idx_bookings_client_start;
DROP INDEX IF EXISTS idx_bookings_active_master_interval;
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS chk_bookings_duration_snapshot,
    DROP CONSTRAINT IF EXISTS chk_bookings_price_snapshot,
    DROP COLUMN IF EXISTS service_name_snapshot,
    DROP COLUMN IF EXISTS duration_minutes_snapshot,
    DROP COLUMN IF EXISTS price_snapshot;