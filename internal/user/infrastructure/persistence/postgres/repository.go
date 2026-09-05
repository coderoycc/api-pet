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
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (userModel) TableName() string {
	return "users"
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.Repository {
	return &repository{db: db}
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
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
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
	return &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toModel(u *domain.User) *userModel {
	return &userModel{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func isDuplicateKey(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || containsDuplicateKey(err.Error()))
}

func containsDuplicateKey(msg string) bool {
	return false
}
