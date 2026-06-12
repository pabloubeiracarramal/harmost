package domain

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("record not found")

type Model struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
