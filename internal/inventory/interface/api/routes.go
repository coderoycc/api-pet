package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(router fiber.Router, handler *InventoryHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	inventory := router.Group("/inventory")

	inventory.Use(authMiddleware)

	// Read routes
	inventory.Get("/batches", handler.GetBatches)
	inventory.Get("/critical-stock", handler.GetCriticalStock)
	inventory.Get("/metrics", handler.GetMetrics)
	inventory.Get("/logs", handler.GetLogs)

	// Write / Operation routes
	inventory.Post("/adjust", roleMiddleware, handler.AdjustStock)
	inventory.Post("/quick-adjust", roleMiddleware, handler.QuickAdjust)
	inventory.Post("/inbound", roleMiddleware, handler.RegisterInbound)
	inventory.Post("/outbound", roleMiddleware, handler.RegisterOutbound)
	inventory.Post("/supplier-return", roleMiddleware, handler.RegisterSupplierReturn)
	inventory.Post("/logs/:id/undo", roleMiddleware, handler.UndoLog)
}
