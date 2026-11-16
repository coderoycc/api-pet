package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"api-go/internal/auth/domain"
	userDomain "api-go/internal/user/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
	GetProfile(ctx context.Context, email string) (*domain.UserPayload, error)
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error)
}

type authService struct {
	userRepo  userDomain.Repository
	roleRepo  userDomain.RoleRepository
	jwtSecret []byte
}

func NewAuthService(userRepo userDomain.Repository, roleRepo userDomain.RoleRepository, secret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		jwtSecret: []byte(secret),
	}
}

func (s *authService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.Status == "suspended" {
		return nil, domain.ErrUserSuspended
	}
	if user.Status == "deleted" {
		return nil, domain.ErrUserDeleted
	}

	payload := domain.ToUserPayload(user)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": payload.ID,
		"email":   payload.Email,
		"role":    payload.Role,
		"exp":     time.Now().Add(6 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: tokenString,
		User:        payload,
	}, nil
}

func (s *authService) GetProfile(ctx context.Context, email string) (*domain.UserPayload, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if user.Status == "suspended" {
		return nil, domain.ErrUserSuspended
	}
	if user.Status == "deleted" {
		return nil, domain.ErrUserDeleted
	}

	return domain.ToUserPayload(user), nil
}

func (s *authService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" || strings.TrimSpace(req.Name) == "" {
		return nil, domain.ErrInvalidInput
	}

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	newUser := &userDomain.User{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(req.Name),
		Email:     email,
		Password:  string(hashedPassword),
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	roleName := req.Role
	if roleName == "" {
		roleName = "Administrador"
	}
	if s.roleRepo != nil {
		role, err := s.roleRepo.FindByName(ctx, roleName)
		if err == nil && role != nil {
			newUser.Role = role
		}
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	payload := domain.ToUserPayload(newUser)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": payload.ID,
		"email":   payload.Email,
		"role":    payload.Role,
		"exp":     time.Now().Add(6 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: tokenString,
		User:        payload,
	}, nil
}
