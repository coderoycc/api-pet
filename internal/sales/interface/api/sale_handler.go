package api

import (
	"errors"
	"strconv"

	"api-go/internal/sales/application"
	"api-go/internal/sales/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type SaleHandler struct {
	service *application.SaleService
}

func NewSaleHandler(service *application.SaleService) *SaleHandler {
	return &SaleHandler{service: service}
}

func getUserInfo(c fiber.Ctx) (uuid.UUID, string) {
	userID := uuid.New()
	userName := "Vendedor"

	if userClaims, ok := c.Locals("user").(jwt.MapClaims); ok {
		if idStr, ok := userClaims["user_id"].(string); ok && idStr != "" {
			if parsed, err := uuid.Parse(idStr); err == nil {
				userID = parsed
			}
		}
		if email, ok := userClaims["email"].(string); ok && email != "" {
			userName = email
		}
	}

	return userID, userName
}

func (h *SaleHandler) ProcessSale(c fiber.Ctx) error {
	var dto application.SaleCreateDto
	if err := c.Bind().JSON(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Datos de venta inválidos: " + err.Error()})
	}

	userID, userName := getUserInfo(c)

	res, err := h.service.ProcessSale(c.Context(), &dto, userID, userName)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyCart) || errors.Is(err, domain.ErrInvalidPayment) {
			return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: err.Error()})
		}
		if errors.Is(err, domain.ErrInsufficientStock) {
			return c.Status(fiber.StatusConflict).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(application.BaseResponse{
		Data:    res,
		Message: "Venta registrada exitosamente",
	})
}

func (h *SaleHandler) GetSaleByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "ID de venta inválido"})
	}

	res, err := h.service.GetSaleByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSaleNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Data: res,
	})
}

func (h *SaleHandler) GetSales(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	method := c.Query("paymentMethod", "")

	filters := &domain.SaleFilters{
		Search:        search,
		PaymentMethod: method,
		Page:          page,
		Limit:         limit,
	}

	res, err := h.service.GetSales(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Data: res,
	})
}

func (h *SaleHandler) GetMetrics(c fiber.Ctx) error {
	res, err := h.service.GetMetrics(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Data: res,
	})
}
