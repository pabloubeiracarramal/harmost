package service

import (
	"context"
	"errors"
	"strings"

	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/repository"
	"gorm.io/gorm"
)

type UserService struct {
	db       *gorm.DB
	userRepo domain.UserRepository
	orgRepo  domain.OrgRepository
}

// SignUpOrLogin upserts a user from a GitHub OAuth profile and returns them
// along with their personal org (creating both on first login).
func (s *UserService) SignUpOrLogin(ctx context.Context, profile domain.GitHubProfile) (*domain.User, *domain.Org, error) {
	// Try to find existing user by GitHub ID.
	existing, err := s.userRepo.GetByGitHubID(ctx, profile.GitHubID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, nil, err
	}
	if err == nil {
		// Returning user — update mutable fields in case they changed on GitHub.
		existing.Name = profile.Name
		existing.AvatarURL = profile.AvatarURL
		if err := s.userRepo.Update(ctx, existing); err != nil {
			return nil, nil, err
		}
		org, err := s.personalOrg(ctx, existing.ID)
		return existing, org, err
	}

	// First login — create user + personal org + membership in one transaction.
	var (
		user domain.User
		org  domain.Org
	)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepos := repository.New(tx)

		user = domain.User{
			GitHubID:  profile.GitHubID,
			Email:     profile.Email,
			Name:      profile.Name,
			AvatarURL: profile.AvatarURL,
		}
		if err := txRepos.User.Create(ctx, &user); err != nil {
			return err
		}

		org = domain.Org{
			Slug:     orgSlug(profile.Login),
			Name:     profile.Name,
			Personal: true,
		}
		if err := txRepos.Org.Create(ctx, &org); err != nil {
			return err
		}

		return txRepos.Org.AddMember(ctx, &domain.OrgMember{
			OrgID:  org.ID,
			UserID: user.ID,
			Role:   domain.RoleOwner,
		})
	})
	return &user, &org, err
}

func (s *UserService) personalOrg(ctx context.Context, userID string) (*domain.Org, error) {
	orgs, err := s.orgRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range orgs {
		if orgs[i].Personal {
			return &orgs[i], nil
		}
	}
	return nil, errors.New("personal org not found")
}

func orgSlug(login string) string {
	return strings.ToLower(strings.ReplaceAll(login, ".", "-"))
}
