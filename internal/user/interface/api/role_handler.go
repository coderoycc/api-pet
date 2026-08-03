package api

import (
	"errors"
	"fmt"
	"strconv"

	"api-go/internal/user/application"
	"api-go/internal/user/domain"

	"github.com/gofiber/fiber/v3"
)

type RoleHandler struct {
	roleService application.RoleService
}

func NewRoleHandler(roleService application.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) GetAllRoles(c fiber.Ctx) error {
	roles, err := h.roleService.GetAllRoles(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch roles",
		})
	}

	dtos := make([]application.RoleDto, len(roles))
	for i, r := range roles {
		dtos[i] = toRoleDto(&r)
	}

	return c.JSON(dtos)
}

func (h *RoleHandler) GetRoleByID(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	role, err := h.roleService.GetRoleByID(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Role not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch role",
		})
	}

	return c.JSON(toRoleDto(role))
}

func (h *RoleHandler) CreateRole(c fiber.Ctx) error {
	var req application.CreateRoleDto
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	role, err := h.roleService.CreateRole(c.Context(), &req)
	if err != nil {
		if errors.Is(err, application.ErrInvalidRoleStatus) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Role name already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create role",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(toRoleDto(role))
}

func (h *RoleHandler) UpdateRole(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	var req application.UpdateRoleDto
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	role, err := h.roleService.UpdateRole(c.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Role not found",
			})
		}
		if errors.Is(err, application.ErrInvalidRoleStatus) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Role name already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update role",
		})
	}

	return c.JSON(toRoleDto(role))
}

func (h *RoleHandler) DeleteRole(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	err = h.roleService.DeleteRole(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Role not found",
			})
		}
		if errors.Is(err, application.ErrRoleHasUsers) || err.Error() == "cannot delete system role: Administrador" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete role",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func toRoleDto(role *domain.Role) application.RoleDto {
	return application.RoleDto{
		ID:          fmt.Sprint(role.ID),
		Name:        role.Name,
		Description: role.Description,
		Status:      role.Status,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}
