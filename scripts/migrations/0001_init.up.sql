CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       email TEXT UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       verified BOOLEAN NOT NULL,
                       role TEXT NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
                                token TEXT PRIMARY KEY,
                                user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                                expires_at TIMESTAMPTZ NOT NULL
);