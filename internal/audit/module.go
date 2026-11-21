package audit

import (
	"api-go/internal/audit/application"
	"api-go/internal/audit/domain"
	postgres "api-go/internal/audit/infrastructure/persistence/postgres"
	"api-go/internal/audit/interface/api"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// RegisterAudit inicializa el módulo de auditoría, registra sus rutas de consulta
// y retorna AuditService para que sea inyectado en cualquier otro módulo.
func RegisterAudit(
	router fiber.Router,
	db *gorm.DB,
	authMiddleware fiber.Handler,
	roleMiddleware fiber.Handler,
) domain.AuditService {
	if err := postgres.AutoMigrate(db); err != nil {
		panic("Failed to migrate audit models: " + err.Error())
	}

	repo := postgres.NewAuditRepository(db)
	service := application.NewAuditService(repo)
	handler := api.NewAuditHandler(service)

	api.RegisterRoutes(router, handler, authMiddleware, roleMiddleware)

	return service
}

// ActorMiddleware expone el middleware que propaga el actor autenticado al context.Context.
func ActorMiddleware() fiber.Handler {
	return api.AuditActorMiddleware()
}
