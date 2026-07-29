package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"api-go/internal/user/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:        uuid.New(),
		Name:      req.Name,
		Password:  string(hash),
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil, err
		}
		return nil, err
	}

	return toResponse(user), nil
}

func (s *Service) GetAll(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = *toResponse(&u)
	}
	return resp, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*UserResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return toResponse(user), nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateUserRequest) (*UserResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	user.Name = req.Name
	user.Email = strings.ToLower(strings.TrimSpace(req.Email))
	user.UpdatedAt = time.Now().UTC()

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, domain.ErrInvalidInput
		}
		user.Password = string(hash)
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toResponse(user), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalidInput
	}

	return s.repo.Delete(ctx, uid)
}

func validateCreate(req CreateUserRequest) error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" || req.Password == "" {
		return domain.ErrInvalidInput
	}
	if !strings.Contains(req.Email, "@") {
		return domain.ErrInvalidInput
	}
	if len(req.Password) < 6 {
		return domain.ErrInvalidInput
	}
	return nil
}

func toResponse(u *domain.User) *UserResponse {
	return &UserResponse{
		ID:    u.ID.String(),
		Name:  u.Name,
		Email: u.Email,
	}
}
