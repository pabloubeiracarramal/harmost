-- +goose Up
CREATE TABLE job_logs (
    id        BIGSERIAL   PRIMARY KEY,
    job_id    UUID        NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    line      TEXT        NOT NULL,
    stream    TEXT        NOT NULL DEFAULT 'unspecified' CHECK (stream IN ('unspecified', 'stdout', 'stderr')),
    sequence  BIGINT      NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL
);

-- query pattern: fetch logs for a job ordered by sequence
CREATE INDEX idx_job_logs_job_id_sequence ON job_logs (job_id, sequence);

-- +goose Down
DROP TABLE job_logs;
