-- +goose Up
ALTER TABLE agent_tokens ADD COLUMN agent_id UUID REFERENCES agents(id) ON DELETE SET NULL;
CREATE INDEX idx_agent_tokens_agent_id ON agent_tokens(agent_id);

-- +goose Down
DROP INDEX IF EXISTS idx_agent_tokens_agent_id;
ALTER TABLE agent_tokens DROP COLUMN agent_id;
