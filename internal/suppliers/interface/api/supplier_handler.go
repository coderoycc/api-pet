package api

import (
	"errors"

	"api-go/internal/suppliers/application"
	"api-go/internal/suppliers/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SupplierHandler struct {
	svc application.SupplierService
}

func NewSupplierHandler(svc application.SupplierService) *SupplierHandler {
	return &SupplierHandler{svc: svc}
}

func (h *SupplierHandler) Create(c fiber.Ctx) error {
	var dto application.CreateSupplierDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if dto.Name == "" || dto.TaxID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and TaxID are required",
		})
	}

	supplier, err := h.svc.Create(c.Context(), dto)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierTaxIdExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(supplier)
}

func (h *SupplierHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid UUID format",
		})
	}

	supplier, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(supplier)
}

func (h *SupplierHandler) GetAll(c fiber.Ctx) error {
	suppliers, err := h.svc.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(suppliers)
}

func (h *SupplierHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid UUID format",
		})
	}

	var dto application.UpdateSupplierDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if dto.Name == "" || dto.TaxID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and TaxID are required",
		})
	}

	err = h.svc.Update(c.Context(), id, dto)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, domain.ErrSupplierTaxIdExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Fetch the updated supplier to return it
	updatedSupplier, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching updated supplier",
		})
	}

	return c.JSON(updatedSupplier)
}

func (h *SupplierHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid UUID format",
		})
	}

	err = h.svc.Delete(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
