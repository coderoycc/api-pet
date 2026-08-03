package domain

import (
	"api-go/internal/user/domain"
)

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        *UserPayload `json:"user"`
}

type UserPayload struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func ToUserPayload(u *domain.User) *UserPayload {
	role := ""
	permissions := []string{}
	if u.Role != nil {
		role = u.Role.Name
		for _, p := range u.Role.Permissions {
			permissions = append(permissions, p.Name)
		}
	}

	return &UserPayload{
		ID:          u.ID.String(),
		Email:       u.Email,
		Name:        u.Name,
		Role:        role,
		Permissions: permissions,
	}
}
