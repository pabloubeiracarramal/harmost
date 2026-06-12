package domain

import "context"

type User struct {
	Model
	Email string `gorm:"uniqueIndex;not null"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type UserService interface {
	SignUp(ctx context.Context, email string) (*User, error)
}
