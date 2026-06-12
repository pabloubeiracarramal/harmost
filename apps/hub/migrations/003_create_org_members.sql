-- +goose Up
CREATE TABLE org_members (
    org_id  UUID NOT NULL REFERENCES orgs(id)  ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role    TEXT NOT NULL CHECK (role IN ('owner', 'member', 'viewer')),
    PRIMARY KEY (org_id, user_id)
);

-- +goose Down
DROP TABLE org_members;
