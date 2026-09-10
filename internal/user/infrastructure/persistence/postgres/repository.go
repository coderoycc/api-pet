package postgres

import (
	"context"
	"errors"
	"time"

	"api-go/internal/user/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Password  string    `gorm:"type:varchar(255);not null"`
	RoleID    *uint
	Role      *roleModel `gorm:"foreignKey:RoleID"`
	Status    string     `gorm:"type:varchar(50);default:'active'"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
}

func (userModel) TableName() string {
	return "users"
}

type roleModel struct {
	ID          uint              `gorm:"primaryKey"`
	Name        string            `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string            `gorm:"type:varchar(255)"`
	Status      string            `gorm:"type:varchar(50);default:'active'"`
	CreatedAt   time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP"`
	Permissions []permissionModel `gorm:"many2many:role_permissions;"`
}

func (roleModel) TableName() string {
	return "roles"
}

type permissionModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);uniqueIndex;not null"`
}

func (permissionModel) TableName() string {
	return "permissions"
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.Repository {
	return &repository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&roleModel{},
		&permissionModel{},
		&userModel{},
	)
}

func (r *repository) Create(ctx context.Context, user *domain.User) error {
	model := toModel(user)
	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *repository) Update(ctx context.Context, user *domain.User) error {
	model := toModel(user)
	err := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"name":       model.Name,
		"email":      model.Email,
		"password":   model.Password,
		"updated_at": model.UpdatedAt,
	}).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&userModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).Preload("Role.Permissions").First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).Preload("Role.Permissions").First(&model, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).First(&model, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *repository) FindAll(ctx context.Context) ([]domain.User, error) {
	var models []userModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, len(models))
	for i, m := range models {
		users[i] = *toDomain(&m)
	}
	return users, nil
}

func toDomain(m *userModel) *domain.User {
	u := &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.Role != nil {
		u.Role = &domain.Role{
			ID:          m.Role.ID,
			Name:        m.Role.Name,
			Description: m.Role.Description,
			Status:      m.Role.Status,
			CreatedAt:   m.Role.CreatedAt,
			UpdatedAt:   m.Role.UpdatedAt,
		}
		if len(m.Role.Permissions) > 0 {
			u.Role.Permissions = make([]domain.Permission, len(m.Role.Permissions))
			for i, p := range m.Role.Permissions {
				u.Role.Permissions[i] = domain.Permission{
					ID:   p.ID,
					Name: p.Name,
				}
			}
		}
	}

	return u
}

func toModel(u *domain.User) *userModel {
	m := &userModel{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if u.Role != nil {
		m.RoleID = &u.Role.ID
	}
	return m
}

func isDuplicateKey(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || containsDuplicateKey(err.Error()))
}

func containsDuplicateKey(msg string) bool {
	return false
}
