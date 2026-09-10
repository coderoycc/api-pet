package api

import (
	"errors"

	"api-go/internal/auth/application"
	"api-go/internal/auth/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
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
	if err := c.Bind().Body(&req); err != nil {
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
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}
		if errors.Is(err, domain.ErrUserSuspended) || errors.Is(err, domain.ErrUserDeleted) {
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
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

func (h *Handler) Profile(c fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	email, ok := userClaims["email"].(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	profile, err := h.authService.GetProfile(c.Context(), email)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		if errors.Is(err, domain.ErrUserSuspended) || errors.Is(err, domain.ErrUserDeleted) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(profile)
}
