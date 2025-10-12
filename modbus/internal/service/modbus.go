package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/goburrow/modbus"

	"modbus-server/internal/config"
	"modbus-server/internal/models"
)

// ModbusService предоставляет методы для работы с Modbus устройствами
type ModbusService struct {
	config *config.Config
}

// NewModbusService создает новый экземпляр ModbusService
func NewModbusService(config *config.Config) *ModbusService {
	return &ModbusService{
		config: config,
	}
}

// WriteCoil записывает значение в coil Modbus устройства
func (s *ModbusService) WriteCoil(req *models.WriteCoilRequest) *models.WriteCoilResponse {
	log.Printf("WriteCoil - box_id: %s, register: %s, value: %v", req.BoxID, req.Register, req.Value)

	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		log.Printf("Modbus отключен, пропускаем запись в регистр %s - box_id: %s, register: %s, value: %v",
			req.Register, req.BoxID, req.Register, req.Value)
		return &models.WriteCoilResponse{
			Success: true,
			Message: "Modbus отключен в конфигурации",
		}
	}

	// Проверяем формат регистра
	if !s.isValidHexRegister(req.Register) {
		log.Printf("Неверный формат регистра - box_id: %s, register: %s", req.BoxID, req.Register)
		return &models.WriteCoilResponse{
			Success: false,
			Message: fmt.Sprintf("Неверный формат регистра: %s", req.Register),
		}
	}

	// Создаем Modbus клиент
	client, handler, err := s.createModbusClient()
	if err != nil {
		log.Printf("Ошибка создания Modbus клиента - box_id: %s, error: %v", req.BoxID, err)
		return &models.WriteCoilResponse{
			Success: false,
			Message: fmt.Sprintf("Не удалось создать Modbus клиент: %v", err),
		}
	}
	defer handler.Close()

	// Конвертируем hex регистр в uint16
	address, err := s.hexToUint16(req.Register)
	if err != nil {
		log.Printf("Ошибка конвертации регистра - box_id: %s, register: %s, error: %v", req.BoxID, req.Register, err)
		return &models.WriteCoilResponse{
			Success: false,
			Message: fmt.Sprintf("Неверный формат регистра: %v", err),
		}
	}

	// Выполняем запись с retry механизмом
	var lastError error
	for attempt := 1; attempt <= 3; attempt++ {
		// Записываем значение в coil
		_, err := client.WriteSingleCoil(address, uint16(boolToUint16(req.Value)))
		if err == nil {
			log.Printf("Modbus: успешно записано значение %v в регистр %s - box_id: %s, register: %s, value: %v",
				req.Value, req.Register, req.BoxID, req.Register, req.Value)
			return &models.WriteCoilResponse{
				Success: true,
				Message: fmt.Sprintf("Успешно записано значение %v в регистр %s", req.Value, req.Register),
			}
		}

		lastError = err
		log.Printf("Modbus: попытка %d неудачна - box_id: %s, register: %s, attempt: %d, error: %v",
			attempt, req.BoxID, req.Register, attempt, err)

		if attempt < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("Modbus: все попытки неудачны - box_id: %s, register: %s, error: %v", req.BoxID, req.Register, lastError)
	return &models.WriteCoilResponse{
		Success: false,
		Message: fmt.Sprintf("Не удалось записать в Modbus после 3 попыток: %v", lastError),
	}
}

// WriteLightCoil включает или выключает свет для бокса
func (s *ModbusService) WriteLightCoil(req *models.WriteLightCoilRequest) *models.WriteLightCoilResponse {
	log.Printf("WriteLightCoil - box_id: %s, value: %v", req.BoxID, req.Value)

	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		log.Printf("Modbus отключен, пропускаем управление светом - box_id: %s, value: %v", req.BoxID, req.Value)
		return &models.WriteLightCoilResponse{
			Success: true,
			Message: "Modbus отключен в конфигурации",
		}
	}

	// Для управления светом нужен регистр, но в упрощенной версии
	// мы будем использовать фиксированный регистр или получать его из конфигурации
	// Пока используем регистр по умолчанию для света
	defaultLightRegister := "0x001"

	coilReq := &models.WriteCoilRequest{
		BoxID:    req.BoxID,
		Register: defaultLightRegister,
		Value:    req.Value,
	}

	coilResp := s.WriteCoil(coilReq)

	return &models.WriteLightCoilResponse{
		Success: coilResp.Success,
		Message: coilResp.Message,
	}
}

// WriteChemistryCoil включает или выключает химию для бокса
func (s *ModbusService) WriteChemistryCoil(req *models.WriteChemistryCoilRequest) *models.WriteChemistryCoilResponse {
	log.Printf("WriteChemistryCoil - box_id: %s, value: %v", req.BoxID, req.Value)

	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		log.Printf("Modbus отключен, пропускаем управление химией - box_id: %s, value: %v", req.BoxID, req.Value)
		return &models.WriteChemistryCoilResponse{
			Success: true,
			Message: "Modbus отключен в конфигурации",
		}
	}

	// Для управления химией используем фиксированный регистр
	defaultChemistryRegister := "0x002"

	coilReq := &models.WriteCoilRequest{
		BoxID:    req.BoxID,
		Register: defaultChemistryRegister,
		Value:    req.Value,
	}

	coilResp := s.WriteCoil(coilReq)

	return &models.WriteChemistryCoilResponse{
		Success: coilResp.Success,
		Message: coilResp.Message,
	}
}

// TestConnection тестирует соединение с Modbus устройством
func (s *ModbusService) TestConnection(req *models.TestConnectionRequest) *models.TestConnectionResponse {
	log.Printf("TestConnection - box_id: %s", req.BoxID)

	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		log.Printf("Modbus отключен в конфигурации - box_id: %s", req.BoxID)
		return &models.TestConnectionResponse{
			Success: false,
			Message: "Modbus протокол отключен в конфигурации",
		}
	}

	// Создаем Modbus клиент
	client, handler, err := s.createModbusClient()
	if err != nil {
		log.Printf("Ошибка создания Modbus клиента для теста - box_id: %s, error: %v", req.BoxID, err)
		return &models.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Не удалось создать Modbus клиент: %v", err),
		}
	}
	defer handler.Close()

	// Пытаемся прочитать регистр для проверки соединения
	_, err = client.ReadCoils(1, 1)
	if err != nil {
		log.Printf("Ошибка чтения coil для теста - box_id: %s, error: %v", req.BoxID, err)
		return &models.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Не удалось подключиться к Modbus устройству: %v", err),
		}
	}

	log.Printf("Тест подключения Modbus успешен - box_id: %s", req.BoxID)
	return &models.TestConnectionResponse{
		Success: true,
		Message: "Соединение с Modbus устройством установлено успешно",
	}
}

// TestCoil тестирует запись в конкретный регистр
func (s *ModbusService) TestCoil(req *models.TestCoilRequest) *models.TestCoilResponse {
	log.Printf("TestCoil - box_id: %s, register: %s, value: %v", req.BoxID, req.Register, req.Value)

	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		log.Printf("Modbus отключен для теста coil - box_id: %s, register: %s", req.BoxID, req.Register)
		return &models.TestCoilResponse{
			Success: false,
			Message: "Modbus протокол отключен в конфигурации",
		}
	}

	// Проверяем формат регистра
	if !s.isValidHexRegister(req.Register) {
		log.Printf("Неверный формат регистра для теста - box_id: %s, register: %s", req.BoxID, req.Register)
		return &models.TestCoilResponse{
			Success: false,
			Message: fmt.Sprintf("Неверный формат регистра: %s", req.Register),
		}
	}

	// Тестируем запись
	coilReq := &models.WriteCoilRequest{
		BoxID:    req.BoxID,
		Register: req.Register,
		Value:    req.Value,
	}

	coilResp := s.WriteCoil(coilReq)

	return &models.TestCoilResponse{
		Success: coilResp.Success,
		Message: coilResp.Message,
	}
}

// createModbusClient создает Modbus TCP клиент
func (s *ModbusService) createModbusClient() (modbus.Client, *modbus.TCPClientHandler, error) {
	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		return nil, nil, fmt.Errorf("Modbus протокол отключен в конфигурации")
	}

	if s.config.ModbusHost == "" {
		return nil, nil, fmt.Errorf("MODBUS_HOST не задан в конфигурации")
	}

	// Создаем TCP клиент с таймаутами для реального ПЛК
	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", s.config.ModbusHost, s.config.ModbusPort))
	handler.Timeout = 10 * time.Second
	handler.SlaveId = 1

	client := modbus.NewClient(handler)
	return client, handler, nil
}

// isValidHexRegister проверяет, что регистр в правильном hex формате
func (s *ModbusService) isValidHexRegister(register string) bool {
	register = strings.ToLower(register)
	if !strings.HasPrefix(register, "0x") {
		return false
	}

	// Проверяем, что после 0x идут только hex символы
	hexPart := register[2:]
	if len(hexPart) == 0 || len(hexPart) > 4 {
		return false
	}

	for _, char := range hexPart {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}

	return true
}

// hexToUint16 конвертирует hex строку в uint16
func (s *ModbusService) hexToUint16(hexStr string) (uint16, error) {
	hexStr = strings.ToLower(hexStr)
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}

	value, err := strconv.ParseUint(hexStr, 16, 16)
	if err != nil {
		return 0, err
	}

	return uint16(value), nil
}

// boolToUint16 конвертирует bool в uint16 для Modbus
func boolToUint16(value bool) uint16 {
	if value {
		return 0xFF00 // Modbus ON
	}
	return 0x0000 // Modbus OFF
}
