package api

import (
	"errors"

	"api-go/internal/settings/application"
	"api-go/internal/settings/domain"

	"github.com/gofiber/fiber/v3"
)

type WarehouseHandler struct {
	svc *application.WarehouseService
}

func NewWarehouseHandler(svc *application.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{svc: svc}
}

func (h *WarehouseHandler) Create(c fiber.Ctx) error {
	var dto application.CreateWarehouseDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if dto.Name == "" || dto.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and Type are required",
		})
	}

	w, err := h.svc.Create(c.Context(), dto)
	if err != nil {
		if errors.Is(err, domain.ErrWarehouseNameExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(application.MapWarehouseToResponse(w))
}

func (h *WarehouseHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	w, err := h.svc.GetByID(c.Context(), idStr)
	if err != nil {
		if errors.Is(err, domain.ErrWarehouseNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format or internal error",
		})
	}

	return c.JSON(application.MapWarehouseToResponse(w))
}

func (h *WarehouseHandler) GetAll(c fiber.Ctx) error {
	warehouses, err := h.svc.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(application.MapWarehouseToResponseList(warehouses))
}

func (h *WarehouseHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	var dto application.UpdateWarehouseDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if dto.Name == "" || dto.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and Type are required",
		})
	}

	w, err := h.svc.Update(c.Context(), idStr, dto)
	if err != nil {
		if errors.Is(err, domain.ErrWarehouseNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, domain.ErrWarehouseNameExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error or invalid ID",
		})
	}

	return c.JSON(application.MapWarehouseToResponse(w))
}

func (h *WarehouseHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	err := h.svc.Delete(c.Context(), idStr)
	if err != nil {
		if errors.Is(err, domain.ErrWarehouseNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error or invalid ID",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
