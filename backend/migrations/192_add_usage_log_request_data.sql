-- Store the original client request body for the usage-record detail view.
-- BYTEA preserves the exact bytes supplied by the client; no redaction,
-- normalization, or JSON re-encoding is applied.
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS request_data BYTEA;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS request_content_type TEXT;

-- Batch image usage is settled asynchronously, so retain the same original
-- request snapshot on the job until its usage log is created.
ALTER TABLE batch_image_jobs ADD COLUMN IF NOT EXISTS request_data BYTEA;
ALTER TABLE batch_image_jobs ADD COLUMN IF NOT EXISTS request_content_type TEXT;
