package customers

import (
	"api-go/internal/customers/application"
	"api-go/internal/customers/infrastructure/persistence/postgres"
	"api-go/internal/customers/interface/api"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterCustomers(router fiber.Router, db *gorm.DB, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	err := postgres.AutoMigrate(db)
	if err != nil {
		panic(err)
	}

	repo := postgres.NewCustomerRepository(db)
	svc := application.NewCustomerService(repo)
	handler := api.NewCustomerHandler(svc)
	
	api.RegisterCustomerRoutes(router, handler, authMiddleware, roleMiddleware)
}
