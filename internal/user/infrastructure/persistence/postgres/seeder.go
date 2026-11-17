package postgres

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var defaultPermissions = []string{
	"create:products", "read:products", "update:products", "delete:products",
	"create:inventory", "read:inventory", "update:inventory", "delete:inventory",
	"create:purchases", "read:purchases", "update:purchases", "delete:purchases",
	"create:customers", "read:customers", "update:customers", "delete:customers",
	"create:suppliers", "read:suppliers", "update:suppliers", "delete:suppliers",
	"create:settings", "read:settings", "update:settings", "delete:settings",
	"create:users", "read:users", "update:users", "delete:users",
	"create:roles", "read:roles", "update:roles", "delete:roles",
}

// Seed inserta los roles, permisos y usuario administrador inicial si no existen.
func Seed(db *gorm.DB) error {
	// 1. Permisos (usamos Find para no disparar logs de 'record not found' en tabla nueva)
	var permissions []permissionModel
	for _, permName := range defaultPermissions {
		var found []permissionModel
		db.Where("name = ?", permName).Limit(1).Find(&found)
		if len(found) > 0 {
			permissions = append(permissions, found[0])
		} else {
			perm := permissionModel{Name: permName}
			if err := db.Create(&perm).Error; err != nil {
				return err
			}
			permissions = append(permissions, perm)
		}
	}

	// 2. Rol Administrador
	var adminRoles []roleModel
	db.Where("name = ?", "Administrador").Limit(1).Find(&adminRoles)
	var adminRole roleModel
	if len(adminRoles) == 0 {
		adminRole = roleModel{
			Name:        "Administrador",
			Description: "Acceso total al sistema",
			Status:      "active",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Permissions: permissions,
		}
		if err := db.Create(&adminRole).Error; err != nil {
			return err
		}
	} else {
		adminRole = adminRoles[0]
		_ = db.Model(&adminRole).Association("Permissions").Replace(&permissions)
	}

	// Roles adicionales estándar
	var sellerRoles []roleModel
	db.Where("name = ?", "Vendedor").Limit(1).Find(&sellerRoles)
	if len(sellerRoles) == 0 {
		sellerRole := roleModel{
			Name:        "Vendedor",
			Description: "Atención a clientes y ventas",
			Status:      "active",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		_ = db.Create(&sellerRole)
	}

	var warehouseRoles []roleModel
	db.Where("name = ?", "Almacenero").Limit(1).Find(&warehouseRoles)
	if len(warehouseRoles) == 0 {
		warehouseRole := roleModel{
			Name:        "Almacenero",
			Description: "Control de inventario y compras",
			Status:      "active",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		_ = db.Create(&warehouseRole)
	}

	// 3. Usuario Administrador
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@example.com"
	}
	adminEmail = strings.ToLower(strings.TrimSpace(adminEmail))

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "AdminPassword123!"
	}

	var existingUsers []userModel
	db.Where("email = ?", adminEmail).Limit(1).Find(&existingUsers)
	if len(existingUsers) == 0 {
		hashed, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		adminUser := userModel{
			ID:        uuid.New(),
			Name:      "Administrador",
			Email:     adminEmail,
			Password:  string(hashed),
			RoleID:    &adminRole.ID,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}
		log.Printf("[SEED] Usuario administrador creado: %s / %s", adminEmail, adminPassword)
	} else {
		adminUser := existingUsers[0]
		if adminUser.RoleID == nil {
			adminUser.RoleID = &adminRole.ID
			_ = db.Save(&adminUser)
		}
	}

	return nil
}
