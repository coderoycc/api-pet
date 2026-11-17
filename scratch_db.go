package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Product struct {
	ID   string `gorm:"primaryKey"`
	Name string
	SKU  string
}

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:1234@localhost:5432/api_go?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	var count int64
	db.Table("products").Count(&count)
	fmt.Printf("TOTAL_PRODUCTS_IN_DB: %d\n", count)

	var prods []Product
	db.Table("products").Limit(10).Find(&prods)
	for _, p := range prods {
		fmt.Printf("ID: %s, SKU: %s, Name: %s\n", p.ID, p.SKU, p.Name)
	}
}
