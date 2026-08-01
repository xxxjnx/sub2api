ALTER TABLE users
    ADD COLUMN IF NOT EXISTS registration_ip INET;

CREATE INDEX IF NOT EXISTS idx_users_registration_ip
    ON users (registration_ip)
    WHERE registration_ip IS NOT NULL;

CREATE TABLE IF NOT EXISTS blocked_registration_ips (
    id BIGSERIAL PRIMARY KEY,
    ip_address INET NOT NULL UNIQUE,
    reason TEXT NOT NULL DEFAULT '',
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blocked_registration_ips_created_at
    ON blocked_registration_ips (created_at DESC);
