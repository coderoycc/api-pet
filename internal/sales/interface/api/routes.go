package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(router fiber.Router, handler *SaleHandler, authMiddleware fiber.Handler) {
	sales := router.Group("/sales")

	sales.Use(authMiddleware)

	sales.Post("/", handler.ProcessSale)
	sales.Get("/metrics", handler.GetMetrics)
	sales.Get("/", handler.GetSales)
	sales.Get("/:id", handler.GetSaleByID)
}
