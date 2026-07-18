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

func (s *UserService) Get(ctx context.Context, id string) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("Y-m-d"),
	}, nil
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

func (s *UserService) Update(ctx context.Context, id string, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}

func (s *UserService) Delete(ctx context.Context, id string, softOpt ...bool) error {
	soft := true
	if len(softOpt) > 0 {
		soft = softOpt[0]
	}

	return s.repo.Delete(ctx, id, soft)
}
