package service

import (
	"context"
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

// SignUp creates a user, a personal org, and an owner membership in one transaction.
func (s *UserService) SignUp(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepos := repository.New(tx)

		user = domain.User{Email: email}
		if err := txRepos.User.Create(ctx, &user); err != nil {
			return err
		}

		org := domain.Org{
			Slug:     orgSlug(email),
			Name:     email,
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
	return &user, err
}

func orgSlug(email string) string {
	local := strings.SplitN(email, "@", 2)[0]
	return strings.ToLower(strings.ReplaceAll(local, ".", "-"))
}
