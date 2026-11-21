package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"api-go/internal/audit"
	auditPostgres "api-go/internal/audit/infrastructure/persistence/postgres"
	"api-go/internal/auth"
	"api-go/internal/customers"
	customersPostgres "api-go/internal/customers/infrastructure/persistence/postgres"
	"api-go/internal/inventory"
	inventoryPostgres "api-go/internal/inventory/infrastructure/persistence/postgres"
	"api-go/internal/products"
	productsPostgres "api-go/internal/products/infrastructure/persistence/postgres"
	"api-go/internal/purchases"
	purchasesPostgres "api-go/internal/purchases/infrastructure/persistence/postgres"
	"api-go/internal/sales"
	"api-go/internal/settings"
	settingsPostgres "api-go/internal/settings/infrastructure/persistence/postgres"
	"api-go/internal/suppliers"
	suppliersPostgres "api-go/internal/suppliers/infrastructure/persistence/postgres"
	"api-go/internal/user"
	userPostgres "api-go/internal/user/infrastructure/persistence/postgres"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretkey"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/api_go?sslmode=disable"
	}

	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migraciones de todos los módulos
	if err := userPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate user database: %v", err)
	}
	if err := settingsPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate settings database: %v", err)
	}
	if err := suppliersPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate suppliers database: %v", err)
	}
	if err := productsPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate products database: %v", err)
	}
	if err := purchasesPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate purchases database: %v", err)
	}
	if err := customersPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate customers database: %v", err)
	}
	if err := inventoryPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate inventory database: %v", err)
	}
	if err := auditPostgres.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate audit database: %v", err)
	}

	// Seeders iniciales
	if err := userPostgres.Seed(db); err != nil {
		log.Printf("warning: failed to seed user data: %v", err)
	}
	if err := settingsPostgres.Seed(db); err != nil {
		log.Printf("warning: failed to seed settings data: %v", err)
	}
	if err := productsPostgres.Seed(db); err != nil {
		log.Printf("warning: failed to seed products data: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
	}))

	user.Register(app.Group("/users"), db)
	auth.Register(app.Group("/auth"), db, jwtSecret)

	authMiddleware := auth.NewAuthMiddleware(jwtSecret)
	roleMiddleware := auth.RequireRole("Administrador")

	user.RegisterRoles(app, db, authMiddleware, roleMiddleware)

	// Settings configuration
	settingsRoleMiddleware := auth.RequireRole("Administrador") // Configurable, se pueden agregar más roles: auth.RequireRole("Administrador", "Gerente")
	settingsGroup := app.Group("/settings")
	settings.RegisterGeneralSettings(settingsGroup, db, authMiddleware, settingsRoleMiddleware)
	settings.RegisterUnitOfMeasure(settingsGroup, db, authMiddleware, settingsRoleMiddleware)
	settings.RegisterWarehouse(settingsGroup, db, authMiddleware, settingsRoleMiddleware)
	user.RegisterRoles(settingsGroup, db, authMiddleware, settingsRoleMiddleware)

	apiV1 := app.Group("/api/v1")
	auditService := audit.RegisterAudit(apiV1, db, authMiddleware, roleMiddleware)
	suppliers.RegisterSuppliers(apiV1, db, authMiddleware, roleMiddleware)
	productRepo := products.RegisterProducts(apiV1, db, authMiddleware, roleMiddleware)
	inventoryRepo := inventory.RegisterInventory(apiV1, db, authMiddleware, roleMiddleware, productRepo)
	purchases.RegisterPurchases(apiV1, db, authMiddleware, roleMiddleware, productRepo, inventoryRepo, auditService)
	customers.RegisterCustomers(apiV1, db, authMiddleware, roleMiddleware)
	sales.RegisterSales(apiV1, db, authMiddleware, productRepo, inventoryRepo, auditService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	log.Printf("server listening on port %s", port)
	log.Fatal(app.Listen(port))
}
