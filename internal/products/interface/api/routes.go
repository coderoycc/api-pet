package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(router fiber.Router, handler *ProductHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	products := router.Group("/products")

	products.Use(authMiddleware)

	// Admin/Manager routes for product modification
	products.Post("/", roleMiddleware, handler.CreateProduct)
	products.Put("/:id", roleMiddleware, handler.UpdateProduct)
	products.Patch("/:id/status", roleMiddleware, handler.UpdateProductStatus)
	products.Delete("/:id", roleMiddleware, handler.DeleteProduct)

	// Read routes
	products.Get("/expiring", handler.GetExpiringProducts)
	products.Get("/low-stock", handler.GetLowStockProducts)
	products.Get("/", handler.GetProducts)
	products.Get("/:id", handler.GetProductByID)
}
