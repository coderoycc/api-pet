package api

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	router.Post("/", handler.Create)
	router.Get("/", handler.GetAll)
	router.Get("/:id", handler.GetByID)
	router.Put("/:id", handler.Update)
	router.Delete("/:id", handler.Delete)
}
