CREATE TABLE IF NOT EXISTS objects (
    id BIGSERIAL PRIMARY KEY,
    bucket VARCHAR(255) NOT NULL,
    path TEXT NOT NULL,
    volume_id UUID NOT NULL,
    needle_id UUID NOT NULL,
    size_bytes BIGINT NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ --soft deletion
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_objects_bucket_path ON objects (bucket, path);

CREATE INDEX IF NOT EXISTS idx_objects_deleted_at ON objects (deleted_at);
