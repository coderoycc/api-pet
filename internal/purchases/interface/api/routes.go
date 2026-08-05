package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(router fiber.Router, handler *PurchaseHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	purchases := router.Group("/purchases")

	purchases.Use(authMiddleware)

	// Admin/Manager routes for modification
	purchases.Post("/", roleMiddleware, handler.CreatePurchase)
	purchases.Post("/:id/receive", roleMiddleware, handler.ReceivePurchase)
	purchases.Post("/:id/cancel", roleMiddleware, handler.CancelPurchase)

	// Read routes
	purchases.Get("/metrics", handler.GetMetrics)
	purchases.Get("/", handler.GetPurchases)
	purchases.Get("/:id", handler.GetPurchaseByID)
}
