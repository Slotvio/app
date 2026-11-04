CREATE TABLE IF NOT EXISTS users (
    id PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT,
    username TEXT,
    phone TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
);