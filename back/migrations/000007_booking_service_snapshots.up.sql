ALTER TABLE bookings
ADD COLUMN service_name_snapshot TEXT,
    ADD COLUMN duration_minutes_snapshot INTEGER,
    ADD COLUMN price_snapshot INTEGER;
UPDATE bookings AS b
SET service_name_snapshot = s.name,
    duration_minutes_snapshot = s.duration_minutes,
    price_snapshot = s.price
FROM services AS s
WHERE s.id = b.service_id;
ALTER TABLE bookings
ALTER COLUMN service_name_snapshot
SET NOT NULL,
    ALTER COLUMN duration_minutes_snapshot
SET NOT NULL,
    ALTER COLUMN price_snapshot
SET NOT NULL,
    ADD CONSTRAINT chk_bookings_duration_snapshot CHECK (
        duration_minutes_snapshot BETWEEN 5 AND 480
    ),
    ADD CONSTRAINT chk_bookings_price_snapshot CHECK (price_snapshot >= 0);
CREATE INDEX idx_bookings_active_master_interval ON bookings (master_id, start_at, end_at)
WHERE status IN ('booked', 'confirmed_by_client');
CREATE INDEX idx_bookings_client_start ON bookings (client_id, start_at);