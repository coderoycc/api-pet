package application

import "time"

type CreateWarehouseDTO struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Location string `json:"location"`
	Type     string `json:"type"`
	Status   bool   `json:"status"`
}

type UpdateWarehouseDTO struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Location string `json:"location"`
	Type     string `json:"type"`
	Status   bool   `json:"status"`
}

type WarehouseResponseDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Location  string    `json:"location"`
	Type      string    `json:"type"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
