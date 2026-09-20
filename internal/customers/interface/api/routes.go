package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterCustomerRoutes(router fiber.Router, handler *CustomerHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	group := router.Group("/customers", authMiddleware)

	// Rutas estáticas PRIMERO para evitar colisión con /:id
	group.Get("/search", handler.Search)
	group.Get("/cities", handler.GetCities)

	// Rutas con parámetro
	group.Get("/", handler.GetCustomers)
	group.Get("/:id", handler.GetByID)

	// Mutaciones (requieren rol)
	group.Post("/", handler.Create, roleMiddleware)
	group.Put("/:id", handler.Update, roleMiddleware)
	group.Delete("/:id", handler.Delete, roleMiddleware)
}
