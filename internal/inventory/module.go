package inventory

import (
	"api-go/internal/inventory/application"
	"api-go/internal/inventory/domain"
	postgres "api-go/internal/inventory/infrastructure/persistence/postgres"
	"api-go/internal/inventory/interface/api"
	productDomain "api-go/internal/products/domain"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterInventory(
	router fiber.Router,
	db *gorm.DB,
	authMiddleware fiber.Handler,
	roleMiddleware fiber.Handler,
	productRepo productDomain.ProductRepository,
) domain.InventoryRepository {
	err := postgres.AutoMigrate(db)
	if err != nil {
		panic("Failed to migrate inventory models: " + err.Error())
	}

	repo := postgres.NewInventoryRepository(db)
	service := application.NewInventoryService(repo, productRepo)
	handler := api.NewInventoryHandler(service)

	api.RegisterRoutes(router, handler, authMiddleware, roleMiddleware)

	return repo
}
