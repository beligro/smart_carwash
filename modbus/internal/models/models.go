package models

import "github.com/google/uuid"

// WriteCoilRequest запрос на запись в coil
type WriteCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id" binding:"required"`
	Register string    `json:"register" binding:"required"`
	Value    bool      `json:"value"`
}

// WriteCoilResponse ответ на запись в coil
type WriteCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// WriteLightCoilRequest запрос на управление светом
type WriteLightCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id" binding:"required"`
	Register string    `json:"register" binding:"required"`
	Value    bool      `json:"value"`
}

// WriteLightCoilResponse ответ на управление светом
type WriteLightCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// WriteChemistryCoilRequest запрос на управление химией
type WriteChemistryCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id" binding:"required"`
	Register string    `json:"register" binding:"required"`
	Value    bool      `json:"value"`
}

// WriteChemistryCoilResponse ответ на управление химией
type WriteChemistryCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestConnectionRequest запрос на тестирование соединения
type TestConnectionRequest struct {
	BoxID uuid.UUID `json:"box_id" binding:"required"`
}

// TestConnectionResponse ответ на тестирование соединения
type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestCoilRequest запрос на тестирование coil
type TestCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id" binding:"required"`
	Register string    `json:"register" binding:"required"`
	Value    bool      `json:"value"`
}

// TestCoilResponse ответ на тестирование coil
type TestCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// BoxConfigRequest запрос с конфигурацией бокса
type BoxConfigRequest struct {
	BoxID                  uuid.UUID `json:"box_id" binding:"required"`
	LightCoilRegister      *string   `json:"light_coil_register,omitempty"`
	ChemistryCoilRegister  *string   `json:"chemistry_coil_register,omitempty"`
}

// BoxConfigResponse ответ с конфигурацией бокса
type BoxConfigResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
