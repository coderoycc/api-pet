package suppliers

import (
	"api-go/internal/suppliers/application"
	"api-go/internal/suppliers/infrastructure/persistence/postgres"
	"api-go/internal/suppliers/interface/api"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterSuppliers(router fiber.Router, db *gorm.DB, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	err := postgres.AutoMigrate(db)
	if err != nil {
		panic(err) // Idealmente manejar el error en main, pero seguimos el estándar
	}

	repo := postgres.NewSupplierRepository(db)
	svc := application.NewSupplierService(repo)
	handler := api.NewSupplierHandler(svc)
	
	api.RegisterSupplierRoutes(router, handler, authMiddleware, roleMiddleware)
}
