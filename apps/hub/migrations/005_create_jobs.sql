-- +goose Up
CREATE TABLE jobs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id)   ON DELETE CASCADE,
    agent_id    UUID        NOT NULL REFERENCES agents(id),
    state       TEXT        NOT NULL DEFAULT 'accepted' CHECK (state IN (
                    'unspecified', 'accepted', 'pulling_image', 'creating_container',
                    'starting_container', 'running', 'stopping', 'cancelled',
                    'succeeded', 'failed', 'timed_out'
                )),
    spec        JSONB       NOT NULL DEFAULT '{}',
    message     TEXT        NOT NULL DEFAULT '',
    exit_code   INTEGER,
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_jobs_org_id   ON jobs (org_id);
CREATE INDEX idx_jobs_agent_id ON jobs (agent_id);

-- +goose Down
DROP TABLE jobs;
