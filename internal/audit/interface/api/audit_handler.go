package api

import (
	"strconv"
	"time"

	"api-go/internal/audit/domain"

	"github.com/gofiber/fiber/v3"
)

type AuditHandler struct {
	service domain.AuditService
}

func NewAuditHandler(service domain.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) ListAuditLogs(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filter := domain.AuditFilter{
		Module:     c.Query("module"),
		Action:     c.Query("action"),
		EntityType: c.Query("entity_type"),
		EntityID:   c.Query("entity_id"),
		ActorID:    c.Query("actor_id"),
		Page:       page,
		Limit:      limit,
	}

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := parseQueryTime(fromStr); err == nil {
			filter.From = &t
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid 'from' date format (expected RFC3339 or YYYY-MM-DD)",
			})
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if t, err := parseQueryTime(toStr); err == nil {
			filter.To = &t
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid 'to' date format (expected RFC3339 or YYYY-MM-DD)",
			})
		}
	}

	res, err := h.service.List(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    res,
		"message": "Audit logs retrieved successfully",
	})
}

func parseQueryTime(val string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", val)
}
