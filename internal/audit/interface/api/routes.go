package api

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, handler *AuditHandler, authMiddleware, roleMiddleware fiber.Handler) {
	group := router.Group("/audit-logs", authMiddleware, roleMiddleware)
	group.Get("/", handler.ListAuditLogs)
}
