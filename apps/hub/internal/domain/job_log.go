package domain

import (
	"context"
	"time"
)

type LogStream string

const (
	LogStreamUnspecified LogStream = "unspecified"
	LogStreamStdout      LogStream = "stdout"
	LogStreamStderr      LogStream = "stderr"
)

type JobLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	JobID     string    `gorm:"type:uuid;not null;index"`
	Line      string    `gorm:"not null"`
	Stream    LogStream `gorm:"type:text;not null"`
	Sequence  int64     `gorm:"not null"`
	Timestamp time.Time `gorm:"not null"`

	Job Job `gorm:"foreignKey:JobID"`
}

type JobLogRepository interface {
	CreateBatch(ctx context.Context, logs []JobLog) error
	ListByJob(ctx context.Context, jobID string) ([]JobLog, error)
}

type JobLogService interface {
	IngestChunks(ctx context.Context, chunks []JobLog) error
	ListByJob(ctx context.Context, jobID string) ([]JobLog, error)
}
