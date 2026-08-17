-- +goose Up
ALTER TABLE agents ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_agents_deleted_at ON agents (deleted_at);

-- +goose Down
DROP INDEX IF EXISTS idx_agents_deleted_at;
ALTER TABLE agents DROP COLUMN deleted_at;
