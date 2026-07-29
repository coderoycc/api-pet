package api

import (
	"errors"
	"net/http"

	"api-go/internal/user/application"
	"api-go/internal/user/domain"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c fiber.Ctx) error {
	var req application.CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resp, err := h.service.Create(c.Context(), req)
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(resp)
}

func (h *Handler) GetAll(c fiber.Ctx) error {
	resp, err := h.service.GetAll(c.Context())
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(http.StatusOK).JSON(resp)
}

func (h *Handler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")

	resp, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(http.StatusOK).JSON(resp)
}

func (h *Handler) Update(c fiber.Ctx) error {
	id := c.Params("id")

	var req application.UpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resp, err := h.service.Update(c.Context(), id, req)
	if err != nil {
		return mapError(c, err)
	}

	return c.Status(http.StatusOK).JSON(resp)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id := c.Params("id")

	err := h.service.Delete(c.Context(), id)
	if err != nil {
		return mapError(c, err)
	}

	return c.SendStatus(http.StatusNoContent)
}

func mapError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	case errors.Is(err, domain.ErrAlreadyExists):
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	case errors.Is(err, domain.ErrInvalidInput):
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	default:
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}
}
