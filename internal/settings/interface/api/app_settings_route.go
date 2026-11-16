package api

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterAppSettingsRoutes(
	router fiber.Router,
	handler *AppSettingsHandler,
	authMiddleware fiber.Handler,
	roleMiddleware fiber.Handler,
) {
	group := router.Group("/general")

	// Lectura: Accesible para cualquier usuario autenticado (para que vendedores también tengan el símbolo de moneda)
	group.Get("/", authMiddleware, handler.Get)

	// Escritura: Solo administradores pueden modificar la configuración general
	group.Put("/", authMiddleware, roleMiddleware, handler.Update)
}
