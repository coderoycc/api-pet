package settings

import (
	"api-go/internal/settings/application"
	"api-go/internal/settings/infrastructure/persistence/postgres"
	"api-go/internal/settings/interface/api"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterUnitOfMeasure(router fiber.Router, db *gorm.DB, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	repo := postgres.NewUnitOfMeasureRepository(db)
	svc := application.NewUnitOfMeasureService(repo)
	handler := api.NewUnitOfMeasureHandler(svc)
	api.RegisterUnitOfMeasureRoutes(router, handler, authMiddleware, roleMiddleware)
}

func RegisterWarehouse(router fiber.Router, db *gorm.DB, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	repo := postgres.NewWarehouseRepository(db)
	svc := application.NewWarehouseService(repo)
	handler := api.NewWarehouseHandler(svc)
	api.RegisterWarehouseRoutes(router, handler, authMiddleware, roleMiddleware)
}
