package purchases

import (
	auditDomain "api-go/internal/audit/domain"
	inventoryDomain "api-go/internal/inventory/domain"
	productDomain "api-go/internal/products/domain"
	"api-go/internal/purchases/application"
	postgres "api-go/internal/purchases/infrastructure/persistence/postgres"
	"api-go/internal/purchases/interface/api"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterPurchases(
	router fiber.Router,
	db *gorm.DB,
	authMiddleware fiber.Handler,
	roleMiddleware fiber.Handler,
	productRepo productDomain.ProductRepository,
	inventoryRepo inventoryDomain.InventoryRepository,
	auditService auditDomain.AuditService,
) {
	err := postgres.AutoMigrate(db)
	if err != nil {
		panic("Failed to migrate purchases models: " + err.Error())
	}

	repo := postgres.NewPurchaseRepository(db)
	service := application.NewPurchaseService(repo, productRepo, inventoryRepo, auditService, db)
	handler := api.NewPurchaseHandler(service)

	api.RegisterRoutes(router, handler, authMiddleware, roleMiddleware)
}
