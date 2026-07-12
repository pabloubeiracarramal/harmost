-- +goose Up
ALTER TABLE users
    ADD COLUMN github_id   TEXT UNIQUE,
    ADD COLUMN name        TEXT NOT NULL DEFAULT '',
    ADD COLUMN avatar_url  TEXT NOT NULL DEFAULT '';

CREATE TABLE agent_tokens (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name           TEXT        NOT NULL,
    token_hash     TEXT        NOT NULL UNIQUE,
    created_by_id  UUID        NOT NULL REFERENCES users(id),
    last_used_at   TIMESTAMPTZ,
    revoked_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agent_tokens_org_id     ON agent_tokens(org_id);
CREATE INDEX idx_agent_tokens_token_hash ON agent_tokens(token_hash);

CREATE TABLE device_codes (
    device_code TEXT        PRIMARY KEY,
    user_code   TEXT        NOT NULL UNIQUE,
    org_id      UUID        REFERENCES orgs(id) ON DELETE CASCADE,
    token       TEXT,
    expires_at  TIMESTAMPTZ NOT NULL,
    approved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE device_codes;
DROP TABLE agent_tokens;
ALTER TABLE users
    DROP COLUMN IF EXISTS github_id,
    DROP COLUMN IF EXISTS name,
    DROP COLUMN IF EXISTS avatar_url;
