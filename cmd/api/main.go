package main

import (
	"log"
	"os"

	"api-go/internal/auth"
	"api-go/internal/user"

	"github.com/gofiber/fiber/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/api_go?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	app := fiber.New()

	user.Register(app.Group("/users"), db)
	auth.Register(app.Group("/auth"), db, jwtSecret)

	log.Fatal(app.Listen(":3000"))
}
