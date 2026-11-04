CREATE TABLE IF NOT EXISTS masters (
    id PRIMARY KEY,
    user_id INT,
    slug TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9_-]{3,40}$'),
    title TEXT NOT NULL,
    category,
    description,
    timezone,
    subscription_until,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
);