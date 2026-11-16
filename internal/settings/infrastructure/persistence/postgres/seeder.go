package postgres

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Seed inserta el almacén y unidad de medida por defecto si no existen.
func Seed(db *gorm.DB) error {
	// 1. Almacén principal por defecto
	var count int64
	db.Model(&warehouseModel{}).Count(&count)
	if count == 0 {
		now := time.Now().UTC()
		warehouse := warehouseModel{
			ID:        uuid.New(),
			Name:      "Almacén Principal",
			Address:   "Av. Principal 123",
			Location:  "Sede Central",
			Type:      "principal",
			Status:    true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&warehouse).Error; err != nil {
			return err
		}
		log.Println("[SEED] Almacén Principal creado por defecto")
	}

	// 2. Unidades de medida por defecto
	defaultUnits := []struct {
		Detail       string
		Abbreviation string
	}{
		{Detail: "Unidad", Abbreviation: "UND"},
		{Detail: "Kilogramo", Abbreviation: "KG"},
		{Detail: "Gramo", Abbreviation: "G"},
		{Detail: "Litro", Abbreviation: "LTS"},
		{Detail: "Saco", Abbreviation: "SACO"},
		{Detail: "Bolsa", Abbreviation: "BLS"},
		{Detail: "Caja", Abbreviation: "CJA"},
	}

	for _, u := range defaultUnits {
		var found int64
		db.Model(&unitOfMeasureModel{}).Where("abbreviation = ?", u.Abbreviation).Count(&found)
		if found == 0 {
			now := time.Now().UTC()
			uom := unitOfMeasureModel{
				ID:           uuid.New(),
				Detail:       u.Detail,
				Abbreviation: u.Abbreviation,
				Status:       "active",
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			_ = db.Create(&uom).Error
		}
	}

	// 3. Configuración general por defecto (Moneda Bs, etc.)
	var settingsCount int64
	db.Model(&appSettingsModel{}).Count(&settingsCount)
	if settingsCount == 0 {
		now := time.Now().UTC()
		defaultSettings := appSettingsModel{
			ID:               uuid.New(),
			CurrencySymbol:   "Bs",
			CurrencyCode:     "BOB",
			CurrencyPosition: "suffix",
			DecimalPlaces:    2,
			CompanyName:      "Pet Shop & Veterinaria",
			CompanyTaxID:     "",
			CompanyPhone:     "",
			CompanyAddress:   "",
			UpdatedAt:        now,
		}
		if err := db.Create(&defaultSettings).Error; err != nil {
			return err
		}
		log.Println("[SEED] Configuración general (Moneda: Bs) creada por defecto")
	}

	return nil
}
