package products

import (
	"api-go/internal/products/application"
	postgres "api-go/internal/products/infrastructure/persistence/postgres"
	"api-go/internal/products/interface/api"

	"api-go/internal/products/domain"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterProducts(router fiber.Router, db *gorm.DB, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) domain.ProductRepository {
	err := postgres.AutoMigrate(db)
	if err != nil {
		panic("Failed to migrate products models: " + err.Error())
	}

	repo := postgres.NewProductRepository(db)
	service := application.NewProductService(repo)
	handler := api.NewProductHandler(service)

	api.RegisterRoutes(router, handler, authMiddleware, roleMiddleware)

	return repo
}
