package main

import (
	"log"
	"os"

	"api-go/internal/auth"
	authHttp "api-go/internal/auth/interface/http"
	"api-go/internal/user"
	userPostgres "api-go/internal/user/infrastructure/persistence/postgres"

	"github.com/gofiber/fiber/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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

	app := fiber.New()

	userRepo := userPostgres.NewRepository(db)

	auth.Register(app.Group("/auth"), userRepo)
	user.Register(app.Group("/users"), db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretkey"
	}
	authMiddleware := authHttp.NewAuthMiddleware(jwtSecret)
	roleMiddleware := authHttp.RequireRole("Administrador")

	user.RegisterRoles(app, db, authMiddleware, roleMiddleware)

	log.Fatal(app.Listen(":3000"))
}
