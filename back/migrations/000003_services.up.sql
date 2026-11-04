CREATE TABLE IF NOT EXISTS services (
    id PRIMARY KEY,
    master_id INT,
    name TEXT NOT NULL,
    description TEXT,
    duration_minutes INT NOT NULL CHECK (
        duration_minutes >= 5
        AND duration_minutes <= 480
    ),
    price INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (master_id) REFERENCES masters (id) ON DELETE CASCADE,
);