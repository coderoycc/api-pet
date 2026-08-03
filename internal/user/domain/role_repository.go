package domain

import "context"

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*Role, error)
	FindAll(ctx context.Context) ([]Role, error)
	HasUsersAssigned(ctx context.Context, id uint) (bool, error)
}
