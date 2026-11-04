CREATE TABLE IF NOT EXISTS bookings (
    id PRIMARY KEY,
    master_id INT,
    client_id INT,
    service_id INT,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status TEXT CHECK (
        status IN (
            'booked',
            'confirmed_by_client',
            'canceled_by_client',
            'canceled_by_master',
            'completed',
            'no_show'
        )
    ),
    client_comment TEXT,
    reminder_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (master_id) REFERENCES masters (id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (service_id) REFERENCES services (id) ON DELETE CASCADE,
);