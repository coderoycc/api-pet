package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterWarehouseRoutes(router fiber.Router, handler *WarehouseHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	wGroup := router.Group("/warehouses", authMiddleware, roleMiddleware)

	wGroup.Post("/", handler.Create)
	wGroup.Get("/", handler.GetAll)
	wGroup.Get("/:id", handler.GetByID)
	wGroup.Put("/:id", handler.Update)
	wGroup.Delete("/:id", handler.Delete)
}
