package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "postgres://postgres:1234@localhost:5432/api_go?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	var productCount int64
	db.Table("products").Count(&productCount)
	fmt.Printf("PRODUCT_COUNT: %d\n", productCount)

	var supplierCount int64
	db.Table("suppliers").Count(&supplierCount)
	fmt.Printf("SUPPLIER_COUNT: %d\n", supplierCount)

	var warehouseCount int64
	db.Table("warehouses").Count(&warehouseCount)
	fmt.Printf("WAREHOUSE_COUNT: %d\n", warehouseCount)

	type ProdSample struct {
		ID   string
		Name string
		SKU  string
	}
	var sample []ProdSample
	db.Table("products").Limit(5).Find(&sample)
	for _, p := range sample {
		fmt.Printf("Product: %s (%s) ID: %s\n", p.Name, p.SKU, p.ID)
	}
}
