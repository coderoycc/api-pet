package api

import (
	"errors"
	"strconv"
	"strings"

	"api-go/internal/customers/application"
	"api-go/internal/customers/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CustomerHandler struct {
	service *application.CustomerService
}

func NewCustomerHandler(service *application.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

// Create crea un nuevo cliente.
// POST /api/v1/customers
func (h *CustomerHandler) Create(c fiber.Ctx) error {
	var dto application.CustomerCreateDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	customer, err := h.service.CreateCustomer(c.Context(), dto)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerDocumentExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"statusCode": fiber.StatusConflict,
				"message":    "Ya existe un cliente registrado con el número de documento " + dto.DocumentNumber,
			})
		}
		if errors.Is(err, domain.ErrInvalidCustomerDocumentType) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"statusCode": fiber.StatusBadRequest,
				"message":    err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(customer)
}

// GetByID obtiene un cliente por su ID.
// GET /api/v1/customers/:id
func (h *CustomerHandler) GetByID(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid customer ID format",
		})
	}

	customer, err := h.service.GetCustomerByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Cliente no encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve customer",
		})
	}

	return c.JSON(customer)
}

// GetCustomers retorna el listado paginado de clientes con filtros y ordenamiento.
// GET /api/v1/customers
func (h *CustomerHandler) GetCustomers(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	// Soporta tanto "limit" como "pageSize"
	limitStr := c.Query("limit", c.Query("pageSize", "10"))
	limit, _ := strconv.Atoi(limitStr)

	sortBy := c.Query("sortBy", "createdAt")

	// Soporta tanto "descending" (bool) como "sortOrder" (asc/desc)
	descending := false
	if d := c.Query("descending"); d == "true" {
		descending = true
	} else if so := c.Query("sortOrder"); strings.ToLower(so) == "desc" {
		descending = true
	}

	filters := &domain.CustomerFilters{
		Search: strings.TrimSpace(c.Query("search")),
		City:   strings.TrimSpace(c.Query("city")),
	}

	// Soporta documentTypes como CSV o parámetros múltiples
	if dtRaw := c.Query("documentTypes"); dtRaw != "" {
		for _, dt := range strings.Split(dtRaw, ",") {
			dt = strings.TrimSpace(dt)
			if dt != "" {
				filters.DocumentTypes = append(filters.DocumentTypes, dt)
			}
		}
	}

	// Soporta statuses como CSV o parámetros múltiples
	if stRaw := c.Query("statuses"); stRaw != "" {
		for _, st := range strings.Split(stRaw, ",") {
			st = strings.TrimSpace(st)
			if st != "" {
				filters.Statuses = append(filters.Statuses, st)
			}
		}
	}

	res, err := h.service.GetCustomers(c.Context(), page, limit, filters, sortBy, descending)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve customers",
		})
	}

	return c.JSON(res)
}

// Search retorna una lista ligera de clientes para autocompletado en selectores.
// GET /api/v1/customers/search
func (h *CustomerHandler) Search(c fiber.Ctx) error {
	// Soporta tanto "q" como "search"
	query := strings.TrimSpace(c.Query("q", c.Query("search")))
	limitStr := c.Query("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	status := c.Query("status", "active")

	results, err := h.service.SearchCustomers(c.Context(), query, limit, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to search customers",
		})
	}

	return c.JSON(results)
}

// GetCities retorna la lista de ciudades únicas registradas en clientes.
// GET /api/v1/customers/cities
func (h *CustomerHandler) GetCities(c fiber.Ctx) error {
	cities, err := h.service.GetDistinctCities(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve cities",
		})
	}

	return c.JSON(cities)
}

// Update actualiza un cliente existente.
// PUT /api/v1/customers/:id
func (h *CustomerHandler) Update(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid customer ID format",
		})
	}

	var dto application.CustomerUpdateDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	customer, err := h.service.UpdateCustomer(c.Context(), id, dto)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Cliente no encontrado",
			})
		}
		if errors.Is(err, domain.ErrCustomerDocumentExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"statusCode": fiber.StatusConflict,
				"message":    "Ya existe un cliente registrado con el número de documento " + dto.DocumentNumber,
			})
		}
		if errors.Is(err, domain.ErrInvalidCustomerDocumentType) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"statusCode": fiber.StatusBadRequest,
				"message":    err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    err.Error(),
		})
	}

	return c.JSON(customer)
}

// Delete elimina un cliente por su ID (con verificación de integridad referencial).
// DELETE /api/v1/customers/:id
func (h *CustomerHandler) Delete(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid customer ID format",
		})
	}

	err = h.service.DeleteCustomer(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Cliente no encontrado",
			})
		}
		if errors.Is(err, domain.ErrCustomerHasAssociatedRecords) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"statusCode": fiber.StatusBadRequest,
				"message":    "El cliente tiene registros asociados y no puede ser eliminado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to delete customer",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Cliente eliminado exitosamente",
	})
}
