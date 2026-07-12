package repository

import (
	"context"

	"github.com/harmost/hub/internal/domain"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	return &u, notFound(err)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return &u, notFound(err)
}

func (r *UserRepo) GetByGitHubID(ctx context.Context, githubID string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).First(&u, "github_id = ?", githubID).Error
	return &u, notFound(err)
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
