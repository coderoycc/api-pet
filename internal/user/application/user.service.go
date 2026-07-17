package application

import (
	"api-pet/internal/user/domain"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo     domain.UserRepository
	validate *validator.Validate
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return &UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
