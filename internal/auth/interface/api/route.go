package api

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	router.Post("/login", handler.Login)
	router.Post("/logout", handler.Logout)
	router.Get("/profile", authMiddleware, handler.Profile)
}
