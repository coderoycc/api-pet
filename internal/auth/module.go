package auth

import (
	"api-go/internal/auth/application"

	"api-go/internal/auth/application"
	"api-go/internal/auth/interface/api"
	"api-go/internal/user/infrastructure/persistence/postgres"

	"github.com/gofiber/fiber/v3"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func Register(router fiber.Router, db *gorm.DB, jwtSecret string) {
	repo := postgres.NewRepository(db)
	svc := application.NewAuthService(repo, []byte(jwtSecret))
	handler := api.NewHandler(svc)
	api.RegisterRoutes(router, handler)
}
