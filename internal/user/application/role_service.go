package application

import (
	"context"
	"errors"
	"time"

	"api-go/internal/user/domain"
)

var (
	ErrRoleHasUsers = errors.New("cannot delete role because it is assigned to users")
	ErrInvalidRoleStatus = errors.New("invalid role status, must be active or inactive")
)

type RoleDto struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateRoleDto struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"max=255"`
	Status      string `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateRoleDto struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"max=255"`
	Status      string `json:"status" validate:"required,oneof=active inactive"`
}

type RoleService interface {
	CreateRole(ctx context.Context, req *CreateRoleDto) (*domain.Role, error)
	UpdateRole(ctx context.Context, id uint, req *UpdateRoleDto) (*domain.Role, error)
	DeleteRole(ctx context.Context, id uint) error
	GetRoleByID(ctx context.Context, id uint) (*domain.Role, error)
	GetAllRoles(ctx context.Context) ([]domain.Role, error)
}

type roleService struct {
	roleRepo domain.RoleRepository
}

func NewRoleService(roleRepo domain.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) CreateRole(ctx context.Context, req *CreateRoleDto) (*domain.Role, error) {
	if req.Status != "active" && req.Status != "inactive" {
		return nil, ErrInvalidRoleStatus
	}

	role := &domain.Role{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id uint, req *UpdateRoleDto) (*domain.Role, error) {
	if req.Status != "active" && req.Status != "inactive" {
		return nil, ErrInvalidRoleStatus
	}

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	role.Name = req.Name
	role.Description = req.Description
	role.Status = req.Status
	role.UpdatedAt = time.Now()

	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) DeleteRole(ctx context.Context, id uint) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate immutability for "Administrador" (Optional business rule)
	if role.Name == "Administrador" {
		return errors.New("cannot delete system role: Administrador")
	}

	hasUsers, err := s.roleRepo.HasUsersAssigned(ctx, id)
	if err != nil {
		return err
	}
	if hasUsers {
		return ErrRoleHasUsers
	}

	return s.roleRepo.Delete(ctx, id)
}

func (s *roleService) GetRoleByID(ctx context.Context, id uint) (*domain.Role, error) {
	return s.roleRepo.FindByID(ctx, id)
}

func (s *roleService) GetAllRoles(ctx context.Context) ([]domain.Role, error) {
	return s.roleRepo.FindAll(ctx)
}
