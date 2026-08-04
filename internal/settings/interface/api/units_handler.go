package api

import (
	"errors"

	"api-go/internal/settings/application"
	"api-go/internal/settings/domain"

	"github.com/gofiber/fiber/v3"
)

type UnitOfMeasureHandler struct {
	svc *application.UnitOfMeasureService
}

func NewUnitOfMeasureHandler(svc *application.UnitOfMeasureService) *UnitOfMeasureHandler {
	return &UnitOfMeasureHandler{svc: svc}
}

func (h *UnitOfMeasureHandler) Create(c fiber.Ctx) error {
	var dto application.CreateUnitOfMeasureDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if dto.Detail == "" || dto.Abbreviation == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Detail and abbreviation are required",
		})
	}

	if dto.Status == "" {
		dto.Status = "active"
	}

	uom, err := h.svc.Create(c.Context(), dto)
	if err != nil {
		if errors.Is(err, domain.ErrUnitAbbrevExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(application.MapUnitOfMeasureToResponse(uom))
}

func (h *UnitOfMeasureHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	uom, err := h.svc.GetByID(c.Context(), idStr)
	if err != nil {
		if errors.Is(err, domain.ErrUnitNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format or internal error",
		})
	}

	return c.JSON(application.MapUnitOfMeasureToResponse(uom))
}

func (h *UnitOfMeasureHandler) GetAll(c fiber.Ctx) error {
	uoms, err := h.svc.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(application.MapUnitOfMeasureToResponseList(uoms))
}

func (h *UnitOfMeasureHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	var dto application.UpdateUnitOfMeasureDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if dto.Detail == "" || dto.Abbreviation == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Detail and abbreviation are required",
		})
	}

	uom, err := h.svc.Update(c.Context(), idStr, dto)
	if err != nil {
		if errors.Is(err, domain.ErrUnitNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, domain.ErrUnitAbbrevExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error or invalid ID",
		})
	}

	return c.JSON(application.MapUnitOfMeasureToResponse(uom))
}

func (h *UnitOfMeasureHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	err := h.svc.Delete(c.Context(), idStr)
	if err != nil {
		if errors.Is(err, domain.ErrUnitNotFound) {
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
