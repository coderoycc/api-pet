package auth

import (
	"api-go/internal/auth/application"
	"api-go/internal/auth/interface/api"
	"api-go/internal/user/infrastructure/persistence/postgres"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func Register(router fiber.Router, db *gorm.DB, jwtSecret string) {
	repo := postgres.NewRepository(db)
	svc := application.NewAuthService(repo, jwtSecret)
	handler := api.NewHandler(svc)
	authMiddleware := api.NewAuthMiddleware(jwtSecret)
	api.RegisterRoutes(router, handler, authMiddleware)
}

func NewAuthMiddleware(secret string) fiber.Handler {
	return api.NewAuthMiddleware(secret)
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return api.RequireRole(allowedRoles...)
}
