package api

import (
	"api-go/internal/audit/application"
	"api-go/internal/audit/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// AttachActor extrae los claims del usuario autenticado (JWT), IP y User-Agent,
// y los adjunta al context.Context de la petición para que cualquier servicio pueda auditar
// sin modificar su firma de parámetros.
func AttachActor(c fiber.Ctx) {
	if userClaims, ok := c.Locals("user").(jwt.MapClaims); ok {
		actor := domain.Actor{
			IPAddress: c.IP(),
			UserAgent: c.Get("User-Agent"),
		}
		if id, ok := userClaims["user_id"].(string); ok {
			actor.ID = id
		}
		if email, ok := userClaims["email"].(string); ok {
			actor.Email = email
		}
		if role, ok := userClaims["role"].(string); ok {
			actor.Role = role
		}

		if actor.ID != "" {
			ctx := application.WithActor(c.Context(), actor)
			c.SetContext(ctx)
		}
	}
}

// AuditActorMiddleware es un middleware Fiber opcional que adjunta el actor al context.Context.
func AuditActorMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		AttachActor(c)
		return c.Next()
	}
}
