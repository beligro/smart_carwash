package client

import "github.com/google/uuid"

// WriteCoilRequest запрос на запись в coil
type WriteCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id"`
	Register string    `json:"register"`
	Value    bool      `json:"value"`
}

// WriteCoilResponse ответ на запись в coil
type WriteCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// WriteLightCoilRequest запрос на управление светом
type WriteLightCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id"`
	Register string    `json:"register"`
	Value    bool      `json:"value"`
}

// WriteLightCoilResponse ответ на управление светом
type WriteLightCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// WriteChemistryCoilRequest запрос на управление химией
type WriteChemistryCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id"`
	Register string    `json:"register"`
	Value    bool      `json:"value"`
}

// WriteChemistryCoilResponse ответ на управление химией
type WriteChemistryCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestCoilRequest запрос на тестирование coil
type TestCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id"`
	Register string    `json:"register"`
	Value    bool      `json:"value"`
}

// TestCoilResponse ответ на тестирование coil
type TestCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
