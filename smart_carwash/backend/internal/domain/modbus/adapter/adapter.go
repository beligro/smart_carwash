package adapter

import (
	"carwash_backend/internal/logger"
	"fmt"
	"time"

	"github.com/google/uuid"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/modbus/client"
	"carwash_backend/internal/domain/modbus/models"
	"carwash_backend/internal/domain/modbus/repository"

	"gorm.io/gorm"
)

// ModbusAdapter адаптер для взаимодействия с modbus сервером через HTTP
type ModbusAdapter struct {
	httpClient *client.ModbusHTTPClient
	repository *repository.ModbusRepository
	config     *config.Config
	db         *gorm.DB
}

// NewModbusAdapter создает новый адаптер для modbus
func NewModbusAdapter(config *config.Config, db *gorm.DB) *ModbusAdapter {
	return &ModbusAdapter{
		httpClient: client.NewModbusHTTPClient(config),
		repository: repository.NewModbusRepository(db),
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
func (a *ModbusAdapter) WriteLightCoil(boxID uuid.UUID, register string, value bool) error {
	logger.Printf("ModbusAdapter WriteLightCoil - box_id: %s, register: %s, value: %v", boxID, register, value)
	
	// Выполняем операцию через HTTP клиент
	err := a.httpClient.WriteLightCoil(boxID, register, value)
	
	// Определяем операцию
	operation := "light_off"
	if value {
		operation = "light_on"
	}
	
	// Сохраняем операцию в БД
	modbusOp := &models.ModbusOperation{
		ID:        uuid.New(),
		BoxID:     boxID,
		Operation: operation,
		Register:  register,
		Value:     value,
		Success:   err == nil,
		CreatedAt: time.Now(),
	}
	
	if err != nil {
		modbusOp.Error = err.Error()
	}
	
	// Сохраняем операцию
	if saveErr := a.repository.SaveModbusOperation(modbusOp); saveErr != nil {
		logger.Printf("Ошибка сохранения операции WriteLightCoil - box_id: %s, error: %v", boxID, saveErr)
	}
	
	// Если операция успешна, обновляем статус койла
	if err == nil {
		if updateErr := a.repository.UpdateModbusCoilStatus(boxID, "light", value); updateErr != nil {
			logger.Printf("Ошибка обновления статуса света - box_id: %s, error: %v", boxID, updateErr)
		}
	}
	
	return err
}

// WriteChemistryCoil включает или выключает химию для бокса
func (a *ModbusAdapter) WriteChemistryCoil(boxID uuid.UUID, register string, value bool) error {
	logger.Printf("ModbusAdapter WriteChemistryCoil - box_id: %s, register: %s, value: %v", boxID, register, value)
	
	// Выполняем операцию через HTTP клиент
	err := a.httpClient.WriteChemistryCoil(boxID, register, value)
	
	// Определяем операцию
	operation := "chemistry_off"
	if value {
		operation = "chemistry_on"
	}
	
	// Сохраняем операцию в БД
	modbusOp := &models.ModbusOperation{
		ID:        uuid.New(),
		BoxID:     boxID,
		Operation: operation,
		Register:  register,
		Value:     value,
		Success:   err == nil,
		CreatedAt: time.Now(),
	}
	
	if err != nil {
		modbusOp.Error = err.Error()
	}
	
	// Сохраняем операцию
	if saveErr := a.repository.SaveModbusOperation(modbusOp); saveErr != nil {
		logger.Printf("Ошибка сохранения операции WriteChemistryCoil - box_id: %s, error: %v", boxID, saveErr)
	}
	
	// Если операция успешна, обновляем статус койла
	if err == nil {
		if updateErr := a.repository.UpdateModbusCoilStatus(boxID, "chemistry", value); updateErr != nil {
			logger.Printf("Ошибка обновления статуса химии - box_id: %s, error: %v", boxID, updateErr)
		}
	}
	
	return err
}

// HandleModbusError обрабатывает ошибку Modbus (только логирует, без продления времени)
func (a *ModbusAdapter) HandleModbusError(boxID uuid.UUID, operation string, sessionID uuid.UUID, err error) error {
	logger.Printf("ModbusAdapter error handler - box_id: %s, operation: %s, session_id: %s, error: %v",
		boxID, operation, sessionID, err)

	return err
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
