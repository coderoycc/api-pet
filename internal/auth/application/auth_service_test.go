package application_test

import (
	"context"
	"errors"
	"testing"

	"api-go/internal/auth/application"
	authDomain "api-go/internal/auth/domain"
	userDomain "api-go/internal/user/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	user *userDomain.User
	err  error
}

func (m *mockUserRepo) Create(ctx context.Context, user *userDomain.User) error      { return nil }
func (m *mockUserRepo) Update(ctx context.Context, user *userDomain.User) error      { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error              { return nil }
func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) FindAll(ctx context.Context) ([]userDomain.User, error) {
	return nil, nil
}

func TestAuthService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	secret := "testsecret"

	t.Run("successful login", func(t *testing.T) {
		repo := &mockUserRepo{
			user: &userDomain.User{
				ID:       uuid.New(),
				Name:     "Test Admin",
				Email:    "admin@example.com",
				Password: string(hashedPassword),
				Status:   "active",
				Role: &userDomain.Role{
					Name: "Administrador",
					Permissions: []userDomain.Permission{
						{Name: "create:users"},
					},
				},
			},
		}

		svc := application.NewAuthService(repo, secret)
		res, err := svc.Login(context.Background(), &authDomain.LoginRequest{
			Email:    "admin@example.com",
			Password: "password123",
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.AccessToken == "" {
			t.Fatal("expected access token to be non-empty")
		}

		if res.User.Email != "admin@example.com" {
			t.Fatalf("expected email admin@example.com, got %s", res.User.Email)
		}

		if res.User.Role != "Administrador" {
			t.Fatalf("expected role Administrador, got %s", res.User.Role)
		}

		// Verify JWT claims
		token, err := jwt.Parse(res.AccessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			t.Fatalf("token is not valid: %v", err)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("invalid claims format")
		}

		if claims["role"] != "Administrador" {
			t.Fatalf("expected claim role Administrador, got %v", claims["role"])
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		repo := &mockUserRepo{
			user: &userDomain.User{
				ID:       uuid.New(),
				Name:     "Test Admin",
				Email:    "admin@example.com",
				Password: string(hashedPassword),
				Status:   "active",
			},
		}

		svc := application.NewAuthService(repo, secret)
		_, err := svc.Login(context.Background(), &authDomain.LoginRequest{
			Email:    "admin@example.com",
			Password: "wrongpassword",
		})

		if !errors.Is(err, authDomain.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("suspended user", func(t *testing.T) {
		repo := &mockUserRepo{
			user: &userDomain.User{
				ID:       uuid.New(),
				Name:     "Suspended User",
				Email:    "suspended@example.com",
				Password: string(hashedPassword),
				Status:   "suspended",
			},
		}

		svc := application.NewAuthService(repo, secret)
		_, err := svc.Login(context.Background(), &authDomain.LoginRequest{
			Email:    "suspended@example.com",
			Password: "password123",
		})

		if !errors.Is(err, authDomain.ErrUserSuspended) {
			t.Fatalf("expected ErrUserSuspended, got %v", err)
		}
	})
}

func TestAuthService_GetProfile(t *testing.T) {
	secret := "testsecret"

	t.Run("successful get profile", func(t *testing.T) {
		repo := &mockUserRepo{
			user: &userDomain.User{
				ID:     uuid.New(),
				Name:   "Admin Profile",
				Email:  "admin@example.com",
				Status: "active",
				Role: &userDomain.Role{
					Name: "Administrador",
					Permissions: []userDomain.Permission{
						{Name: "read:users"},
					},
				},
			},
		}

		svc := application.NewAuthService(repo, secret)
		profile, err := svc.GetProfile(context.Background(), "admin@example.com")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if profile.Email != "admin@example.com" {
			t.Fatalf("expected email admin@example.com, got %s", profile.Email)
		}

		if profile.Role != "Administrador" {
			t.Fatalf("expected role Administrador, got %s", profile.Role)
		}
	})

	t.Run("profile user not found", func(t *testing.T) {
		repo := &mockUserRepo{
			err: userDomain.ErrNotFound,
		}

		svc := application.NewAuthService(repo, secret)
		_, err := svc.GetProfile(context.Background(), "unknown@example.com")

		if !errors.Is(err, authDomain.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}
