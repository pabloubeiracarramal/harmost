package service

import (
	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/repository"
	"gorm.io/gorm"
)

// Services bundles all service instances.
type Services struct {
	User   *UserService
	Agent  *AgentService
	Job    *JobService
	JobLog *JobLogService
}

func New(db *gorm.DB) *Services {
	repos := repository.New(db)
	return &Services{
		User:   &UserService{db: db, userRepo: repos.User, orgRepo: repos.Org},
		Agent:  &AgentService{db: db, agentRepo: repos.Agent},
		Job:    &JobService{jobRepo: repos.Job},
		JobLog: &JobLogService{jobLogRepo: repos.JobLog},
	}
}

// compile-time interface checks
var _ domain.UserService = (*UserService)(nil)
var _ domain.AgentService = (*AgentService)(nil)
var _ domain.JobService = (*JobService)(nil)
var _ domain.JobLogService = (*JobLogService)(nil)
