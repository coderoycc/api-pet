package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterCustomerRoutes(router fiber.Router, handler *CustomerHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	group := router.Group("/customers", authMiddleware)

	group.Post("/", handler.Create, roleMiddleware)
	group.Get("/:id", handler.GetByID)
	group.Get("/", handler.GetAll)
	group.Put("/:id", handler.Update, roleMiddleware)
	group.Delete("/:id", handler.Delete, roleMiddleware)
}
