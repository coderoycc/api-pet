package application

import (
	"context"
	"errors"
	"time"

	"api-go/internal/auth/domain"
	userDomain "api-go/internal/user/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserSuspended      = errors.New("user is suspended")
	ErrUserDeleted        = errors.New("user is deleted")
)

type AuthService interface {
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
}

type authService struct {
	userRepo userDomain.Repository
	jwtSecret []byte
}

func NewAuthService(userRepo userDomain.Repository, secret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: []byte(secret),
	}
}

func (s *authService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, userDomain.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status == "suspended" {
		return nil, ErrUserSuspended
	}
	if user.Status == "deleted" {
		return nil, ErrUserDeleted
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
