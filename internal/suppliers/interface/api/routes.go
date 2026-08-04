package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterSupplierRoutes(router fiber.Router, handler *SupplierHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	group := router.Group("/suppliers", authMiddleware)

	// Configuración de rutas
	group.Post("/", handler.Create, roleMiddleware)
	group.Get("/:id", handler.GetByID)
	group.Get("/", handler.GetAll)
	group.Put("/:id", handler.Update, roleMiddleware)
	group.Delete("/:id", handler.Delete, roleMiddleware)
}
