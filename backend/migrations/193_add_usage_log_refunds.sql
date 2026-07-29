ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS refund_reason TEXT,
    ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS refunded_by BIGINT REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_usage_logs_refunded_at
    ON usage_logs (refunded_at)
    WHERE refunded_at IS NOT NULL;
