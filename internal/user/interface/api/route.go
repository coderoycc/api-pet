package api

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	router.Post("/", handler.Create)
	router.Get("/", handler.GetAll)
	router.Get("/:id", handler.GetByID)
	router.Put("/:id", handler.Update)
	router.Delete("/:id", handler.Delete)
}

func RegisterRoleRoutes(router fiber.Router, handler *RoleHandler, authMiddleware fiber.Handler, roleMiddleware fiber.Handler) {
	rolesGroup := router.Group("/roles")
	
	rolesGroup.Use(authMiddleware)
	rolesGroup.Use(roleMiddleware)

	rolesGroup.Post("/", handler.CreateRole)
	rolesGroup.Get("/", handler.GetAllRoles)
	rolesGroup.Get("/:id", handler.GetRoleByID)
	rolesGroup.Put("/:id", handler.UpdateRole)
	rolesGroup.Delete("/:id", handler.DeleteRole)
}
