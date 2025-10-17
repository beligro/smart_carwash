package adapter

import (
	"fmt"
	"carwash_backend/internal/logger"
	"time"

	"github.com/google/uuid"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/modbus/client"
	sessionModels "carwash_backend/internal/domain/session/models"
	"gorm.io/gorm"
)

// ModbusAdapter адаптер для взаимодействия с modbus сервером через HTTP
type ModbusAdapter struct {
	httpClient *client.ModbusHTTPClient
	config     *config.Config
	db         *gorm.DB
}

// NewModbusAdapter создает новый адаптер для modbus
func NewModbusAdapter(config *config.Config, db *gorm.DB) *ModbusAdapter {
	return &ModbusAdapter{
		httpClient: client.NewModbusHTTPClient(config),
		config:     config,
		db:         db,
	}
}

// WriteCoil записывает значение в coil Modbus устройства
func (a *ModbusAdapter) WriteCoil(boxID uuid.UUID, register string, value bool) error {
	logger.Printf("ModbusAdapter WriteCoil - box_id: %s, register: %s, value: %v", boxID, register, value)
	return a.httpClient.WriteCoil(boxID, register, value)
}

// WriteLightCoil включает или выключает свет для бокса
func (a *ModbusAdapter) WriteLightCoil(boxID uuid.UUID, value bool) error {
	logger.Printf("ModbusAdapter WriteLightCoil - box_id: %s, value: %v", boxID, value)
	return a.httpClient.WriteLightCoil(boxID, value)
}

// WriteChemistryCoil включает или выключает химию для бокса
func (a *ModbusAdapter) WriteChemistryCoil(boxID uuid.UUID, value bool) error {
	logger.Printf("ModbusAdapter WriteChemistryCoil - box_id: %s, value: %v", boxID, value)
	return a.httpClient.WriteChemistryCoil(boxID, value)
}

// HandleModbusError обрабатывает ошибку Modbus и продлевает время сессии
func (a *ModbusAdapter) HandleModbusError(boxID uuid.UUID, operation string, sessionID uuid.UUID, err error) error {
	logger.Printf("ModbusAdapter error handler - box_id: %s, operation: %s, session_id: %s, error: %v",
		boxID, operation, sessionID, err)

	// Проверяем, не связана ли ошибка с отключенным Modbus
	if err != nil && (err.Error() == "Modbus протокол отключен в конфигурации" ||
		err.Error() == "Modbus отключен, пропускаем запись в регистр") {
		// Если Modbus отключен, не продлеваем время сессии
		logger.Printf("Modbus отключен, не продлеваем время сессии - session_id: %s", sessionID)
		return err
	}

	// Продлеваем время сессии на 5 минут только для реальных ошибок Modbus
	if err := a.ExtendSessionTime(sessionID, 5*time.Minute); err != nil {
		logger.Printf("Не удалось продлить время сессии - session_id: %s, error: %v", sessionID, err)
	}

	logger.Printf("Modbus ошибка через HTTP клиент - session_id: %s, error: %v", sessionID, err)

	return err
}

// ExtendSessionTime продлевает время сессии
func (a *ModbusAdapter) ExtendSessionTime(sessionID uuid.UUID, duration time.Duration) error {
	var session sessionModels.Session
	if err := a.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return fmt.Errorf("не удалось найти сессию: %v", err)
	}

	// Продлеваем время в зависимости от статуса сессии
	if session.Status == "in_queue" {
		// Для сессий в очереди продлеваем created_at
		newCreatedAt := session.CreatedAt.Add(-duration)
		return a.db.Model(&session).Update("created_at", newCreatedAt).Error
	} else if session.Status == "active" {
		// Для активных сессий продлеваем status_updated_at
		newStatusUpdatedAt := session.StatusUpdatedAt.Add(-duration)
		return a.db.Model(&session).Update("status_updated_at", newStatusUpdatedAt).Error
	}

	return nil
}

// TestConnection тестирует соединение с Modbus устройством
func (a *ModbusAdapter) TestConnection(boxID uuid.UUID) error {
	logger.Printf("ModbusAdapter TestConnection - box_id: %s", boxID)
	
	resp, err := a.httpClient.TestConnection(boxID)
	if err != nil {
		return err
	}
	
	if !resp.Success {
		return fmt.Errorf("тест соединения неудачен: %s", resp.Message)
	}
	
	return nil
}

// TestCoil тестирует запись в конкретный регистр
func (a *ModbusAdapter) TestCoil(boxID uuid.UUID, register string, value bool) error {
	logger.Printf("ModbusAdapter TestCoil - box_id: %s, register: %s, value: %v", boxID, register, value)
	
	resp, err := a.httpClient.TestCoil(boxID, register, value)
	if err != nil {
		return err
	}
	
	if !resp.Success {
		return fmt.Errorf("тест coil неудачен: %s", resp.Message)
	}
	
	return nil
}
