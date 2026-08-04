package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterUnitOfMeasureRoutes(router fiber.Router, handler *UnitOfMeasureHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	uomGroup := router.Group("/units-of-measure", authMiddleware, roleMiddleware)

	uomGroup.Post("/", handler.Create)
	uomGroup.Get("/", handler.GetAll)
	uomGroup.Get("/:id", handler.GetByID)
	uomGroup.Put("/:id", handler.Update)
	uomGroup.Delete("/:id", handler.Delete)
}
