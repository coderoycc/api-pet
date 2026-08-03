package postgres

import (
	"context"
	"errors"
	"api-go/internal/user/domain"

	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	model := toRoleModel(role)
	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	role.ID = model.ID
	role.CreatedAt = model.CreatedAt
	role.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	model := toRoleModel(role)
	err := r.db.WithContext(ctx).Model(&roleModel{}).Where("id = ?", role.ID).Updates(map[string]interface{}{
		"name":        model.Name,
		"description": model.Description,
		"status":      model.Status,
		"updated_at":  model.UpdatedAt,
	}).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&roleModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *roleRepository) FindByID(ctx context.Context, id uint) (*domain.Role, error) {
	var model roleModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toRoleDomain(&model), nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]domain.Role, error) {
	var models []roleModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	roles := make([]domain.Role, len(models))
	for i, m := range models {
		roles[i] = *toRoleDomain(&m)
	}
	return roles, nil
}

func (r *roleRepository) HasUsersAssigned(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userModel{}).Where("role_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func toRoleModel(r *domain.Role) *roleModel {
	return &roleModel{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func toRoleDomain(m *roleModel) *domain.Role {
	return &domain.Role{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
