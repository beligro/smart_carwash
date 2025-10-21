package models

import (
	"time"

	"github.com/google/uuid"
)

// ModbusOperation представляет операцию Modbus
type ModbusOperation struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	BoxID     uuid.UUID `json:"box_id"`
	Operation string    `json:"operation"` // "light_on", "light_off", "chemistry_on", "chemistry_off"
	Register  string    `json:"register"`  // hex формат
	Value     bool      `json:"value"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ModbusConfig представляет конфигурацию Modbus устройства
type ModbusConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ModbusCommand представляет команду для Modbus устройства
type ModbusCommand struct {
	BoxID     uuid.UUID `json:"box_id"`
	Operation string    `json:"operation"`
	Register  string    `json:"register"`
	Value     bool      `json:"value"`
}

// ModbusError представляет ошибку Modbus операции
type ModbusError struct {
	BoxID     uuid.UUID `json:"box_id"`
	Operation string    `json:"operation"`
	Error     string    `json:"error"`
	SessionID uuid.UUID `json:"session_id,omitempty"`
}

// TestModbusCoilRequest запрос на тестирование Modbus coil
type TestModbusCoilRequest struct {
	BoxID    uuid.UUID `json:"box_id" binding:"required"`
	Register string    `json:"register" binding:"required"`
	Value    bool      `json:"value"`
}

// TestModbusCoilResponse ответ на тестирование Modbus coil
type TestModbusCoilResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ModbusStatus представляет статус Modbus устройства
type ModbusStatus struct {
	BoxID     uuid.UUID `json:"box_id"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Connected bool      `json:"connected"`
	LastError string    `json:"last_error,omitempty"`
	LastSeen  time.Time `json:"last_seen,omitempty"`
}

// GetModbusStatusRequest запрос на получение статуса Modbus
type GetModbusStatusRequest struct {
	BoxID uuid.UUID `json:"box_id" binding:"required"`
}

// GetModbusStatusResponse ответ на получение статуса Modbus
type GetModbusStatusResponse struct {
	Status ModbusStatus `json:"status"`
}

// ModbusStats представляет статистику операций Modbus
type ModbusStats struct {
	TotalOperations          int64             `json:"total_operations"`
	SuccessfulOperations     int64             `json:"successful_operations"`
	FailedOperations         int64             `json:"failed_operations"`
	LastOperation            *ModbusOperation  `json:"last_operation,omitempty"`
	LastSuccessfulOperation  *ModbusOperation  `json:"last_successful_operation,omitempty"`
}

// ModbusConnectionStatus представляет статус подключения Modbus для бокса
type ModbusConnectionStatus struct {
	ID              uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	BoxID           uuid.UUID `json:"box_id" gorm:"uniqueIndex"`
	Connected       bool      `json:"connected"`
	LastError       string    `json:"last_error"`
	LastSeen        time.Time `json:"last_seen"`
	LightStatus     *bool     `json:"light_status,omitempty"`     // Статус света: true - включен, false - выключен, nil - неизвестно
	ChemistryStatus *bool     `json:"chemistry_status,omitempty"` // Статус химии: true - включена, false - выключена, nil - неизвестно
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ModbusDashboardData представляет данные для дашборда мониторинга
type ModbusDashboardData struct {
	Overview         ModbusDashboardOverview   `json:"overview"`
	BoxStatuses      []ModbusBoxStatus         `json:"box_statuses"`
	RecentOperations []ModbusOperation         `json:"recent_operations"`
	ErrorStats       map[string]int64          `json:"error_stats"`
}

// ModbusDashboardOverview представляет общую статистику
type ModbusDashboardOverview struct {
	TotalBoxes           int   `json:"total_boxes"`
	ConnectedBoxes       int   `json:"connected_boxes"`
	DisconnectedBoxes    int   `json:"disconnected_boxes"`
	OperationsLast24h    int64 `json:"operations_last_24h"`
	SuccessRateLast24h   float64 `json:"success_rate_last_24h"`
}

// ModbusBoxStatus представляет статус конкретного бокса
type ModbusBoxStatus struct {
	BoxID                 uuid.UUID `json:"box_id"`
	BoxNumber             int       `json:"box_number"`
	Connected             bool      `json:"connected"`
	LastError             string    `json:"last_error,omitempty"`
	LightCoilRegister     *string   `json:"light_coil_register,omitempty"`
	ChemistryCoilRegister *string   `json:"chemistry_coil_register,omitempty"`
	LightStatus           *bool     `json:"light_status,omitempty"`     // Статус света: true - включен, false - выключен, nil - неизвестно
	ChemistryStatus       *bool     `json:"chemistry_status,omitempty"` // Статус химии: true - включена, false - выключена, nil - неизвестно
}

// GetModbusDashboardRequest запрос на получение данных дашборда
type GetModbusDashboardRequest struct {
	TimeRange string `json:"time_range" form:"time_range"` // "1h", "24h", "7d", "30d"
}

// GetModbusDashboardResponse ответ с данными дашборда
type GetModbusDashboardResponse struct {
	Data ModbusDashboardData `json:"data"`
}

// GetModbusHistoryRequest запрос на получение истории операций
type GetModbusHistoryRequest struct {
	BoxID     *uuid.UUID `json:"box_id" form:"box_id"`
	Operation *string    `json:"operation" form:"operation"`
	Success   *bool      `json:"success" form:"success"`
	Limit     int        `json:"limit" form:"limit"`
	Offset    int        `json:"offset" form:"offset"`
	Since     *time.Time `json:"since" form:"since"`
}

// GetModbusHistoryResponse ответ с историей операций
type GetModbusHistoryResponse struct {
	Operations []ModbusOperation `json:"operations"`
	Total      int64             `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}
