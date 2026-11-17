package api

import (
	"api-go/internal/settings/application"

	"github.com/gofiber/fiber/v3"
)

type AppSettingsHandler struct {
	svc application.AppSettingsService
}

func NewAppSettingsHandler(svc application.AppSettingsService) *AppSettingsHandler {
	return &AppSettingsHandler{svc: svc}
}

func (h *AppSettingsHandler) Get(c fiber.Ctx) error {
	settings, err := h.svc.GetSettings(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al obtener la configuración del sistema",
		})
	}

	return c.JSON(settings)
}

func (h *AppSettingsHandler) Update(c fiber.Ctx) error {
	var dto application.UpdateAppSettingsDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cuerpo de solicitud inválido",
		})
	}

	settings, err := h.svc.UpdateSettings(c.Context(), dto)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al actualizar la configuración del sistema",
		})
	}

	return c.JSON(settings)
}
