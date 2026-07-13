package domain

import "context"

type Org struct {
	Model
	Slug     string `gorm:"uniqueIndex;not null"`
	Name     string `gorm:"not null"`
	Personal bool   `gorm:"not null;default:false"`
}

type OrgMemberRole string

const (
	RoleOwner  OrgMemberRole = "owner"
	RoleMember OrgMemberRole = "member"
	RoleViewer OrgMemberRole = "viewer"
)

type OrgMember struct {
	OrgID  string        `gorm:"type:uuid;primaryKey"`
	UserID string        `gorm:"type:uuid;primaryKey"`
	Role   OrgMemberRole `gorm:"type:text;not null"`

	Org  Org  `gorm:"foreignKey:OrgID"`
	User User `gorm:"foreignKey:UserID"`
}

type OrgRepository interface {
	Create(ctx context.Context, org *Org) error
	GetByID(ctx context.Context, id string) (*Org, error)
	GetBySlug(ctx context.Context, slug string) (*Org, error)
	ListByUserID(ctx context.Context, userID string) ([]Org, error)
	AddMember(ctx context.Context, m *OrgMember) error
	GetMembership(ctx context.Context, orgID, userID string) (*OrgMember, error)
}
