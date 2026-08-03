package http

import (
	"errors"

	"api-go/internal/auth/application"
	"api-go/internal/auth/domain"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	authService application.AuthService
}

func NewHandler(authService application.AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	res, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}
		if errors.Is(err, application.ErrUserSuspended) || errors.Is(err, application.ErrUserDeleted) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *Handler) Logout(c fiber.Ctx) error {
	// In a stateless JWT implementation, logout is usually handled client-side
	// by discarding the token.
	// For server-side logout, we could implement a token blacklist here.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}
