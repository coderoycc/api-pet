package api

import (
	"strconv"

	"api-go/internal/inventory/application"
	"api-go/internal/inventory/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	service application.InventoryService
}

func NewInventoryHandler(service application.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

func getUserInfo(c fiber.Ctx) (string, string) {
	userID := "1"
	userName := "Sistema"

	if id, ok := c.Locals("user_id").(string); ok && id != "" {
		userID = id
	}
	if name, ok := c.Locals("user_name").(string); ok && name != "" {
		userName = name
	}

	return userID, userName
}

func (h *InventoryHandler) GetBatches(c fiber.Ctx) error {
	productIDStr := c.Query("productId")
	var productID *uuid.UUID

	if productIDStr != "" {
		parsed, err := uuid.Parse(productIDStr)
		if err == nil {
			productID = &parsed
		}
	}

	batches, err := h.service.GetBatches(c.Context(), productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(batches)
}

func (h *InventoryHandler) GetCriticalStock(c fiber.Ctx) error {
	minStock := 5
	if minStr := c.Query("minStock"); minStr != "" {
		if val, err := strconv.Atoi(minStr); err == nil {
			minStock = val
		}
	}

	items, err := h.service.GetCriticalStock(c.Context(), minStock)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *InventoryHandler) GetMetrics(c fiber.Ctx) error {
	metrics, err := h.service.GetMetrics(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(metrics)
}

func (h *InventoryHandler) GetLogs(c fiber.Ctx) error {
	filters := &domain.InventoryFilters{}

	if t := c.Query("type"); t != "" {
		logType := domain.InventoryLogType(t)
		filters.Type = &logType
	}
	if pStr := c.Query("productId"); pStr != "" {
		if pID, err := uuid.Parse(pStr); err == nil {
			filters.ProductID = &pID
		}
	}
	if bStr := c.Query("batchId"); bStr != "" {
		if bID, err := uuid.Parse(bStr); err == nil {
			filters.BatchID = &bID
		}
	}
	if uStr := c.Query("userId"); uStr != "" {
		filters.UserID = &uStr
	}
	if rStr := c.Query("reasonType"); rStr != "" {
		filters.ReasonType = &rStr
	}
	if search := c.Query("search"); search != "" {
		filters.Search = search
	}
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			filters.Page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = l
		}
	}
	if sortBy := c.Query("sortBy"); sortBy != "" {
		filters.SortBy = sortBy
	}
	if sortOrder := c.Query("sortOrder"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	result, err := h.service.GetLogs(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

func (h *InventoryHandler) AdjustStock(c fiber.Ctx) error {
	var req domain.AdjustStockRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos: " + err.Error()})
	}

	userID, userName := getUserInfo(c)
	log, adjustment, err := h.service.AdjustStock(c.Context(), req, userID, userName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"log":        log,
		"adjustment": adjustment,
	})
}

func (h *InventoryHandler) QuickAdjust(c fiber.Ctx) error {
	var adjustments []domain.QuickAdjustment
	if err := c.Bind().Body(&adjustments); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos: " + err.Error()})
	}

	userID, userName := getUserInfo(c)
	logs, err := h.service.QuickAdjust(c.Context(), adjustments, userID, userName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(logs)
}

func (h *InventoryHandler) RegisterInbound(c fiber.Ctx) error {
	var op domain.InboundOperation
	if err := c.Bind().Body(&op); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos: " + err.Error()})
	}

	userID, userName := getUserInfo(c)
	log, err := h.service.RegisterInbound(c.Context(), op, userID, userName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(log)
}

func (h *InventoryHandler) RegisterOutbound(c fiber.Ctx) error {
	var op domain.OutboundOperation
	if err := c.Bind().Body(&op); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos: " + err.Error()})
	}

	userID, userName := getUserInfo(c)
	logs, err := h.service.RegisterOutbound(c.Context(), op, userID, userName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(logs)
}

func (h *InventoryHandler) RegisterSupplierReturn(c fiber.Ctx) error {
	var op domain.SupplierReturnOperation
	if err := c.Bind().Body(&op); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Datos inválidos: " + err.Error()})
	}

	userID, userName := getUserInfo(c)
	logs, err := h.service.RegisterSupplierReturn(c.Context(), op, userID, userName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(logs)
}

func (h *InventoryHandler) UndoLog(c fiber.Ctx) error {
	logIDStr := c.Params("id")
	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID de registro inválido"})
	}

	log, err := h.service.UndoLog(c.Context(), logID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(log)
}
