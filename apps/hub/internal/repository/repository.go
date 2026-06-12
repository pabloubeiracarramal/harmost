package repository

import (
	"errors"

	"github.com/harmost/hub/internal/domain"
	"gorm.io/gorm"
)

type Repos struct {
	User   *UserRepo
	Org    *OrgRepo
	Agent  *AgentRepo
	Job    *JobRepo
	JobLog *JobLogRepo
}

func New(db *gorm.DB) *Repos {
	return &Repos{
		User:   &UserRepo{db: db},
		Org:    &OrgRepo{db: db},
		Agent:  &AgentRepo{db: db},
		Job:    &JobRepo{db: db},
		JobLog: &JobLogRepo{db: db},
	}
}

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	return err
}
