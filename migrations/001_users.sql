CREATE TABLE IF NOT EXISTS users
(
    id         SERIAL PRIMARY KEY,
    login      TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

