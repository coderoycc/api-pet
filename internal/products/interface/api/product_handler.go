package api

import (
	"errors"
	"strconv"

	"api-go/internal/products/application"
	"api-go/internal/products/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service *application.ProductService
}

func NewProductHandler(service *application.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) CreateProduct(c fiber.Ctx) error {
	var dto application.ProductCreateDto
	if err := c.Bind().JSON(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid input data"})
	}

	res, err := h.service.CreateProduct(c.Context(), &dto)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateSKU) {
			return c.Status(fiber.StatusConflict).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(application.BaseResponse{
		Data:    res,
		Message: "Product created successfully",
	})
}

func (h *ProductHandler) UpdateProduct(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid product ID"})
	}

	var dto application.ProductUpdateDto
	if err := c.Bind().JSON(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid input data"})
	}
	dto.ID = id.String()

	res, err := h.service.UpdateProduct(c.Context(), id, &dto)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(application.BaseResponse{Message: err.Error()})
		}
		if errors.Is(err, domain.ErrDuplicateSKU) {
			return c.Status(fiber.StatusConflict).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Data:    res,
		Message: "Product updated successfully",
	})
}

func (h *ProductHandler) GetProductByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid product ID"})
	}

	res, err := h.service.GetProductByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Data:    res,
		Message: "Product retrieved successfully",
	})
}

func (h *ProductHandler) GetProducts(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "created_at")
	sortOrder := c.Query("sortOrder", "desc")

	filters := &domain.ProductFilters{
		Search:   c.Query("search"),
		Category: c.Query("category"),
		Status:   c.Query("status"),
	}

	if minPrice := c.Query("minPrice"); minPrice != "" {
		if val, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filters.MinPrice = &val
		}
	}
	if maxPrice := c.Query("maxPrice"); maxPrice != "" {
		if val, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filters.MaxPrice = &val
		}
	}
	if supplierID := c.Query("supplierId"); supplierID != "" {
		if id, err := uuid.Parse(supplierID); err == nil {
			filters.SupplierID = &id
		}
	}
	
	// Para tags, podríamos recibirlos como "tag1,tag2"
	if tags := c.Query("tags"); tags != "" {
		// Asumiendo que tags viene en CSV o array
		// Aquí lo tomamos como simple, depende de cómo lo envíe el frontend.
		filters.Tags = append(filters.Tags, tags)
	}

	res, err := h.service.GetProducts(c.Context(), page, limit, filters, sortBy, sortOrder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *ProductHandler) DeleteProduct(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid product ID"})
	}

	err = h.service.DeleteProduct(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Message: "Product deleted successfully",
	})
}

func (h *ProductHandler) UpdateProductStatus(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid product ID"})
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(application.BaseResponse{Message: "Invalid input data"})
	}

	err = h.service.UpdateProductStatus(c.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(application.BaseResponse{Message: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(application.BaseResponse{
		Message: "Product status updated successfully",
	})
}

func (h *ProductHandler) GetExpiringProducts(c fiber.Ctx) error {
	filters := &domain.ExpiringProductFilters{
		Search:   c.Query("search"),
		Category: c.Query("category"),
		Urgency:  c.Query("urgency"),
	}

	if daysAhead := c.Query("daysAhead"); daysAhead != "" {
		if val, err := strconv.Atoi(daysAhead); err == nil {
			filters.DaysAhead = &val
		}
	}

	res, err := h.service.GetExpiringProducts(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *ProductHandler) GetLowStockProducts(c fiber.Ctx) error {
	threshold := 10
	if thStr := c.Query("threshold"); thStr != "" {
		if val, err := strconv.Atoi(thStr); err == nil {
			threshold = val
		}
	}

	res, err := h.service.GetLowStockProducts(c.Context(), threshold)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(application.BaseResponse{Message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
