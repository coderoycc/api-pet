package api

import (
	"errors"

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

func (h *CustomerHandler) Create(c fiber.Ctx) error {
	var dto application.CustomerCreateDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if dto.Name == "" || dto.DocumentType == "" || dto.DocumentNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, DocumentType and DocumentNumber are required",
		})
	}

	customer, err := h.service.CreateCustomer(c.Context(), dto)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerDocumentExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Customer with this document number already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create customer",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(customer)
}

func (h *CustomerHandler) GetByID(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid customer ID format",
		})
	}

	customer, err := h.service.GetCustomerByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Customer not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve customer",
		})
	}

	return c.JSON(customer)
}

func (h *CustomerHandler) GetAll(c fiber.Ctx) error {
	customers, err := h.service.GetAllCustomers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve customers",
		})
	}

	return c.JSON(customers)
}

func (h *CustomerHandler) Update(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid customer ID format",
		})
	}

	var dto application.CustomerUpdateDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if dto.Name == "" || dto.DocumentType == "" || dto.DocumentNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, DocumentType and DocumentNumber are required",
		})
	}

	customer, err := h.service.UpdateCustomer(c.Context(), id, dto)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Customer not found",
			})
		}
		if errors.Is(err, domain.ErrCustomerDocumentExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Customer with this document number already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update customer",
		})
	}

	return c.JSON(customer)
}

func (h *CustomerHandler) Delete(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid customer ID format",
		})
	}

	err = h.service.DeleteCustomer(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Customer not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete customer",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
