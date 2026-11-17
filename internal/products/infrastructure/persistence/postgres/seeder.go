package postgres

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Seed inserta productos iniciales por defecto si no existen o sus presentaciones si faltan.
func Seed(db *gorm.DB) error {
	now := time.Now().UTC()
	prodAID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	prodBID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	var count int64
	db.Model(&productModel{}).Count(&count)
	if count == 0 {
		prods := []productModel{
			{
				ID:          prodAID,
				SKU:         "SKU-001",
				Name:        "Alimento Canino Premium",
				Description: "Alimento balanceado completo para perros adultos",
				Category:    "Alimentos",
				Subcategory: "Perros",
				Price:       25.00,
				Cost:        15.00,
				Quantity:    50.0,
				MinQuantity: 10.0,
				Unit:        "kg",
				Status:      "active",
				CreatedAt:   now,
				UpdatedAt:   now,
				Presentations: []productPresentationModel{
					{
						ID:               uuid.MustParse("11111111-aaaa-1111-1111-111111111111"),
						ProductID:        prodAID,
						Name:             "Kilo Suelto",
						Barcode:          "SKU-001-KG",
						ConversionFactor: 1.0,
						SalePrice:        25.00,
						IsDefault:        true,
						CreatedAt:        now,
						UpdatedAt:        now,
					},
					{
						ID:               uuid.MustParse("11111111-bbbb-1111-1111-111111111111"),
						ProductID:        prodAID,
						Name:             "Saco 15 kg",
						Barcode:          "SKU-001-S15",
						ConversionFactor: 15.0,
						SalePrice:        320.00,
						IsDefault:        false,
						CreatedAt:        now,
						UpdatedAt:        now,
					},
				},
			},
			{
				ID:          prodBID,
				SKU:         "SKU-002",
				Name:        "Antiparasitario NexGard Spectra",
				Description: "Tableta masticable antiparasitaria para perros",
				Category:    "Farmacia",
				Subcategory: "Salud",
				Price:       45.00,
				Cost:        30.00,
				Quantity:    20.0,
				MinQuantity: 5.0,
				Unit:        "und",
				Status:      "active",
				CreatedAt:   now,
				UpdatedAt:   now,
				Presentations: []productPresentationModel{
					{
						ID:               uuid.MustParse("22222222-aaaa-2222-2222-222222222222"),
						ProductID:        prodBID,
						Name:             "Caja x 1 Tableta",
						Barcode:          "SKU-002-UND",
						ConversionFactor: 1.0,
						SalePrice:        45.00,
						IsDefault:        true,
						CreatedAt:        now,
						UpdatedAt:        now,
					},
				},
			},
		}

		for _, p := range prods {
			if err := db.Create(&p).Error; err != nil {
				log.Printf("[SEED] warning creating initial product %s: %v", p.Name, err)
			} else {
				log.Printf("[SEED] Producto inicial creado: %s (%s)", p.Name, p.ID)
			}
		}
	} else {
		// Asegurar que existan presentaciones para los productos ya existentes
		var presCount int64
		db.Model(&productPresentationModel{}).Count(&presCount)
		if presCount == 0 {
			var prods []productModel
			db.Find(&prods)
			for _, p := range prods {
				uName := p.Unit
				if uName == "" {
					uName = "Unidad"
				}
				pres := productPresentationModel{
					ID:               uuid.New(),
					ProductID:        p.ID,
					Name:             uName,
					Barcode:          p.SKU,
					ConversionFactor: 1.0,
					SalePrice:        p.Price,
					IsDefault:        true,
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				db.Create(&pres)
			}
			log.Printf("[SEED] Presentaciones por defecto creadas para productos existentes.")
		}
	}

	return nil
}
