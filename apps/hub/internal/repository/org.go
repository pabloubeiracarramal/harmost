package repository

import (
	"context"

	"github.com/harmost/hub/internal/domain"
	"gorm.io/gorm"
)

type OrgRepo struct {
	db *gorm.DB
}

func (r *OrgRepo) Create(ctx context.Context, org *domain.Org) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *OrgRepo) GetByID(ctx context.Context, id string) (*domain.Org, error) {
	var o domain.Org
	err := r.db.WithContext(ctx).First(&o, "id = ?", id).Error
	return &o, notFound(err)
}

func (r *OrgRepo) GetBySlug(ctx context.Context, slug string) (*domain.Org, error) {
	var o domain.Org
	err := r.db.WithContext(ctx).First(&o, "slug = ?", slug).Error
	return &o, notFound(err)
}

// ListByUserID returns all orgs the user belongs to.
func (r *OrgRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Org, error) {
	var orgs []domain.Org
	err := r.db.WithContext(ctx).
		Joins("JOIN org_members ON org_members.org_id = orgs.id").
		Where("org_members.user_id = ?", userID).
		Find(&orgs).Error
	return orgs, err
}

func (r *OrgRepo) AddMember(ctx context.Context, m *domain.OrgMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *OrgRepo) GetMembership(ctx context.Context, orgID, userID string) (*domain.OrgMember, error) {
	var m domain.OrgMember
	err := r.db.WithContext(ctx).
		First(&m, "org_id = ? AND user_id = ?", orgID, userID).Error
	return &m, notFound(err)
}
