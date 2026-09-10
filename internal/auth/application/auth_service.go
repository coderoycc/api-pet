package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"api-go/internal/auth/domain"
	userDomain "api-go/internal/user/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
	GetProfile(ctx context.Context, email string) (*domain.UserPayload, error)
}

type authService struct {
	userRepo  userDomain.Repository
	jwtSecret []byte
}

func NewAuthService(userRepo userDomain.Repository, secret string) AuthService {
	return &authService{
		userRepo:  userRepo,
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
