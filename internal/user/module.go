package user

import (
	"api-go/internal/user/application"
	"api-go/internal/user/infrastructure/persistence/postgres"
	"api-go/internal/user/interface/api"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func Register(router fiber.Router, db *gorm.DB) {
	repo := postgres.NewRepository(db)
	svc := application.NewService(repo)
	handler := api.NewHandler(svc)
	api.RegisterRoutes(router, handler)
}
