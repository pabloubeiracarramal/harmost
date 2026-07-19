package service

import (
	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/events"
	"github.com/harmost/hub/internal/repository"
	"gorm.io/gorm"
)

type Services struct {
	User        *UserService
	Agent       *AgentService
	Job         *JobService
	JobLog      *JobLogService
	AgentToken  *AgentTokenService
	DeviceFlow  *DeviceFlowService
}

func New(db *gorm.DB, frontendURL string, bus *events.Bus) *Services {
	repos := repository.New(db)
	agentTokenSvc := &AgentTokenService{db: db, tokenRepo: repos.AgentToken}
	return &Services{
		User:   &UserService{db: db, userRepo: repos.User, orgRepo: repos.Org},
		Agent:  &AgentService{db: db, agentRepo: repos.Agent},
		Job:    &JobService{jobRepo: repos.Job, bus: bus},
		JobLog: &JobLogService{jobLogRepo: repos.JobLog},
		AgentToken: agentTokenSvc,
		DeviceFlow: &DeviceFlowService{
			db:          db,
			deviceRepo:  repos.DeviceCode,
			tokenRepo:   repos.AgentToken,
			tokenSvc:    agentTokenSvc,
			frontendURL: frontendURL,
		},
	}
}

// compile-time interface checks
var _ domain.UserService = (*UserService)(nil)
var _ domain.AgentService = (*AgentService)(nil)
var _ domain.JobService = (*JobService)(nil)
var _ domain.JobLogService = (*JobLogService)(nil)
var _ domain.AgentTokenService = (*AgentTokenService)(nil)
var _ domain.DeviceFlowService = (*DeviceFlowService)(nil)
