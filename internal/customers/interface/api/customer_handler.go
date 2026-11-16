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
		return respondError(c, fiber.StatusBadRequest, "Invalid request payload")
	}

	customer, err := h.service.CreateCustomer(c.Context(), dto)
	if err != nil {
		return handleCustomerError(c, err, dto.DocumentNumber)
	}

	return c.Status(fiber.StatusCreated).JSON(customer)
}

// GetByID obtiene un cliente por su ID.
// GET /api/v1/customers/:id
func (h *CustomerHandler) GetByID(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "Invalid customer ID format")
	}

	customer, err := h.service.GetCustomerByID(c.Context(), id)
	if err != nil {
		return handleCustomerError(c, err, "")
	}

	return c.JSON(customer)
}

// GetCustomers retorna el listado paginado de clientes con filtros y ordenamiento.
// GET /api/v1/customers
func (h *CustomerHandler) GetCustomers(c fiber.Ctx) error {
	page, limit, filters, sortBy, descending := parseCustomerListQuery(c)

	res, err := h.service.GetCustomers(c.Context(), page, limit, filters, sortBy, descending)
	if err != nil {
		return respondError(c, fiber.StatusInternalServerError, "Failed to retrieve customers")
	}

	return c.JSON(res)
}

// Search retorna una lista ligera de clientes para autocompletado en selectores.
// GET /api/v1/customers/search
func (h *CustomerHandler) Search(c fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q", c.Query("search")))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	status := c.Query("status", domain.StatusActive)

	results, err := h.service.SearchCustomers(c.Context(), query, limit, status)
	if err != nil {
		return respondError(c, fiber.StatusInternalServerError, "Failed to search customers")
	}

	return c.JSON(results)
}

// GetCities retorna la lista de ciudades únicas registradas en clientes.
// GET /api/v1/customers/cities
func (h *CustomerHandler) GetCities(c fiber.Ctx) error {
	cities, err := h.service.GetDistinctCities(c.Context())
	if err != nil {
		return respondError(c, fiber.StatusInternalServerError, "Failed to retrieve cities")
	}

	return c.JSON(cities)
}

// Update actualiza un cliente existente.
// PUT /api/v1/customers/:id
func (h *CustomerHandler) Update(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "Invalid customer ID format")
	}

	var dto application.CustomerUpdateDTO
	if err := c.Bind().Body(&dto); err != nil {
		return respondError(c, fiber.StatusBadRequest, "Invalid request payload")
	}

	customer, err := h.service.UpdateCustomer(c.Context(), id, dto)
	if err != nil {
		return handleCustomerError(c, err, dto.DocumentNumber)
	}

	return c.JSON(customer)
}

// Delete elimina un cliente por su ID (con verificación de integridad referencial).
// DELETE /api/v1/customers/:id
func (h *CustomerHandler) Delete(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "Invalid customer ID format")
	}

	if err := h.service.DeleteCustomer(c.Context(), id); err != nil {
		return handleCustomerError(c, err, "")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// --- Helpers HTTP ---

func parseUUIDParam(c fiber.Ctx, param string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(param))
}

func parseCustomerListQuery(c fiber.Ctx) (int, int, *domain.CustomerFilters, string, bool) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limitStr := c.Query("limit", c.Query("pageSize", "10"))
	limit, _ := strconv.Atoi(limitStr)
	sortBy := c.Query("sortBy", "createdAt")

	descending := false
	if c.Query("descending") == "true" || strings.ToLower(c.Query("sortOrder")) == "desc" {
		descending = true
	}

	filters := &domain.CustomerFilters{
		Search:        strings.TrimSpace(c.Query("search")),
		City:          strings.TrimSpace(c.Query("city")),
		DocumentTypes: splitCSV(c.Query("documentTypes")),
		Statuses:      splitCSV(c.Query("statuses")),
	}

	return page, limit, filters, sortBy, descending
}

func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	var items []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func respondError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"statusCode": status,
		"message":    message,
	})
}

func handleCustomerError(c fiber.Ctx, err error, docNumber string) error {
	switch {
	case errors.Is(err, domain.ErrCustomerNotFound):
		return respondError(c, fiber.StatusNotFound, "Cliente no encontrado")
	case errors.Is(err, domain.ErrCustomerDocumentExists):
		msg := "Ya existe un cliente registrado con ese número de documento"
		if docNumber != "" {
			msg = "Ya existe un cliente registrado con el número de documento " + docNumber
		}
		return respondError(c, fiber.StatusConflict, msg)
	case errors.Is(err, domain.ErrCustomerHasAssociatedRecords):
		return respondError(c, fiber.StatusBadRequest, "El cliente tiene registros asociados y no puede ser eliminado")
	case errors.Is(err, domain.ErrCustomerNameRequired),
		errors.Is(err, domain.ErrCustomerDocumentTypeRequired),
		errors.Is(err, domain.ErrCustomerDocumentNumberRequired),
		errors.Is(err, domain.ErrInvalidCustomerDocumentType),
		errors.Is(err, domain.ErrInvalidCustomerStatus),
		errors.Is(err, domain.ErrInvalidCustomerEmail):
		return respondError(c, fiber.StatusBadRequest, err.Error())
	default:
		return respondError(c, fiber.StatusInternalServerError, "Error interno al procesar la solicitud")
	}
}
