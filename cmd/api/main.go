package main

import (
	"log"
	"os"

	"api-go/internal/auth"
	"api-go/internal/customers"
	"api-go/internal/inventory"
	"api-go/internal/products"
	"api-go/internal/purchases"
	"api-go/internal/settings"
	settingsPostgres "api-go/internal/settings/infrastructure/persistence/postgres"
	"api-go/internal/suppliers"
	"api-go/internal/user"
	userPostgres "api-go/internal/user/infrastructure/persistence/postgres"

	"github.com/gofiber/fiber/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretkey"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/api_go?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := userPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	if err := settingsPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	app := fiber.New()

	user.Register(app.Group("/users"), db)
	auth.Register(app.Group("/auth"), db, jwtSecret)

	authMiddleware := auth.NewAuthMiddleware(jwtSecret)
	roleMiddleware := auth.RequireRole("Administrador")

	user.RegisterRoles(app, db, authMiddleware, roleMiddleware)

	// Settings configuration
	settingsRoleMiddleware := auth.RequireRole("Administrador") // Configurable, se pueden agregar más roles: auth.RequireRole("Administrador", "Gerente")
	settingsGroup := app.Group("/settings")
	settings.RegisterUnitOfMeasure(settingsGroup, db, authMiddleware, settingsRoleMiddleware)
	settings.RegisterWarehouse(settingsGroup, db, authMiddleware, settingsRoleMiddleware)

	suppliers.RegisterSuppliers(app.Group("/api/v1"), db, authMiddleware, roleMiddleware)
	productRepo := products.RegisterProducts(app.Group("/api/v1"), db, authMiddleware, roleMiddleware)
	purchases.RegisterPurchases(app.Group("/api/v1"), db, authMiddleware, roleMiddleware, productRepo)
	customers.RegisterCustomers(app.Group("/api/v1"), db, authMiddleware, roleMiddleware)
	inventory.RegisterInventory(app.Group("/api/v1"), db, authMiddleware, roleMiddleware, productRepo)

	log.Fatal(app.Listen(":3000"))
}
