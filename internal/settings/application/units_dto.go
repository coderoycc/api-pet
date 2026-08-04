package application

import "time"

type CreateUnitOfMeasureDTO struct {
	Detail       string `json:"detail"`
	Abbreviation string `json:"abbreviation"`
	Status       string `json:"status"`
}

type UpdateUnitOfMeasureDTO struct {
	Detail       string `json:"detail"`
	Abbreviation string `json:"abbreviation"`
	Status       string `json:"status"`
}

type UnitOfMeasureResponseDTO struct {
	ID           string    `json:"id"`
	Detail       string    `json:"detail"`
	Abbreviation string    `json:"abbreviation"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
