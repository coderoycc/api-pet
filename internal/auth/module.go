package auth

import (
	"os"

	"api-go/internal/auth/application"
	authHttp "api-go/internal/auth/interface/http"
	userDomain "api-go/internal/user/domain"

	"github.com/gofiber/fiber/v3"
)

func Register(router fiber.Router, userRepo userDomain.Repository) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecretkey" // Default for development, should be overridden in production
	}

	authService := application.NewAuthService(userRepo, secret)
	handler := authHttp.NewHandler(authService)

	router.Post("/login", handler.Login)
	router.Post("/logout", handler.Logout)
}
