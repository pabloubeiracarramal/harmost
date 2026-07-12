-- +goose Up
ALTER TABLE agents
    ADD COLUMN cpu_usage_percent   REAL,
    ADD COLUMN memory_used_bytes   BIGINT,
    ADD COLUMN memory_total_bytes  BIGINT,
    ADD COLUMN disk_used_bytes     BIGINT,
    ADD COLUMN disk_total_bytes    BIGINT,
    ADD COLUMN running_containers  INTEGER;

-- +goose Down
ALTER TABLE agents
    DROP COLUMN IF EXISTS cpu_usage_percent,
    DROP COLUMN IF EXISTS memory_used_bytes,
    DROP COLUMN IF EXISTS memory_total_bytes,
    DROP COLUMN IF EXISTS disk_used_bytes,
    DROP COLUMN IF EXISTS disk_total_bytes,
    DROP COLUMN IF EXISTS running_containers;
