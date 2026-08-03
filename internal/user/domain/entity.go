package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Password  string
	Role      *Role
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Role struct {
	ID          uint
	Name        string // admin, inventory_operator, etc
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Permissions []Permission
}

type Permission struct {
	ID   uint
	Name string // create:products, read:products, etc
}
