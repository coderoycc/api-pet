package api

import (
	"api-go/internal/auth/application"
	"api-go/internal/auth/domain"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *application.AuthService
}

func NewHandler(service *application.AuthService) *Handler {
	return &Handler{service}
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req application.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	resp, err := h.service.Login(c.Context(), req)
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(http.StatusOK).JSON(resp)
}

func mapError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	case errors.Is(err, domain.ErrInvalidToken):
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})

	default:
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}
}
