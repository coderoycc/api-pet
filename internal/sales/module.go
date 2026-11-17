package sales

import (
	inventoryDomain "api-go/internal/inventory/domain"
	productDomain "api-go/internal/products/domain"
	"api-go/internal/sales/application"
	postgres "api-go/internal/sales/infrastructure/persistence/postgres"
	"api-go/internal/sales/interface/api"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterSales(
	router fiber.Router,
	db *gorm.DB,
	authMiddleware fiber.Handler,
	productRepo productDomain.ProductRepository,
	inventoryRepo inventoryDomain.InventoryRepository,
) {
	err := postgres.AutoMigrate(db)
	if err != nil {
		panic("Failed to migrate sales models: " + err.Error())
	}

	repo := postgres.NewSaleRepository(db)
	service := application.NewSaleService(repo, productRepo, inventoryRepo, db)
	handler := api.NewSaleHandler(service)

	api.RegisterRoutes(router, handler, authMiddleware)
}
