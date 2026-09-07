package api

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	router.Post("/login", handler.Login)
}
