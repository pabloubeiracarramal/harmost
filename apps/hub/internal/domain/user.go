package domain

import "context"

type User struct {
	Model
	GitHubID  string `gorm:"column:github_id;uniqueIndex"`
	Email     string `gorm:"uniqueIndex;not null"`
	Name      string
	AvatarURL string
}

type GitHubProfile struct {
	GitHubID  string
	Email     string
	Name      string
	AvatarURL string
	Login     string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*User, error)
	Update(ctx context.Context, user *User) error
}

type UserService interface {
	SignUpOrLogin(ctx context.Context, profile GitHubProfile) (*User, *Org, error)
	GetByID(ctx context.Context, id string) (*User, error)
}
