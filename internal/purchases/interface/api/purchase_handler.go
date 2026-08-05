package api

import (
	"errors"
	"strconv"

	"api-go/internal/purchases/application"
	"api-go/internal/purchases/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type PurchaseHandler struct {
	service *application.PurchaseService
}

func NewPurchaseHandler(service *application.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{service: service}
}

func (h *PurchaseHandler) CreatePurchase(c fiber.Ctx) error {
	var dto application.PurchaseCreateDto
	if err := c.Bind().JSON(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid input data"})
	}

	res, err := h.service.CreatePurchase(c.Context(), &dto)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(application.BaseResponse{
		Data:    res,
		Message: "Purchase created successfully",
	})
}

func (h *PurchaseHandler) ReceivePurchase(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid purchase ID"})
	}

	res, err := h.service.ReceivePurchase(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(application.BaseResponse{Message: err.Error()})
		}
		if errors.Is(err, domain.ErrCannotReceiveCancelled) || errors.Is(err, domain.ErrAlreadyReceived) {
			return c.Status(fiber.StatusConflict).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Data:    res,
		Message: "Purchase received successfully",
	})
}

func (h *PurchaseHandler) CancelPurchase(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid purchase ID"})
	}

	res, err := h.service.CancelPurchase(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(application.BaseResponse{Message: err.Error()})
		}
		if errors.Is(err, domain.ErrCannotCancelReceived) {
			return c.Status(fiber.StatusConflict).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Data:    res,
		Message: "Purchase cancelled successfully",
	})
}

func (h *PurchaseHandler) GetPurchaseByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid purchase ID"})
	}

	res, err := h.service.GetPurchaseByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Data:    res,
		Message: "Purchase retrieved successfully",
	})
}

func (h *PurchaseHandler) GetPurchases(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "created_at")
	sortOrder := c.Query("sortOrder", "desc")

	filters := &domain.PurchaseFilters{
		Search:    c.Query("search"),
		Status:    c.Query("status"),
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}

	if supplierID := c.Query("supplierId"); supplierID != "" {
		if id, err := uuid.Parse(supplierID); err == nil {
			filters.SupplierID = &id
		}
	}
	if warehouseID := c.Query("warehouseId"); warehouseID != "" {
		if id, err := uuid.Parse(warehouseID); err == nil {
			filters.WarehouseID = &id
		}
	}

	res, err := h.service.GetPurchases(c.Context(), page, limit, filters, sortBy, sortOrder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *PurchaseHandler) GetMetrics(c fiber.Ctx) error {
	res, err := h.service.GetMetrics(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
