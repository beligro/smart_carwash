package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/goburrow/modbus"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/modbus/models"
	"carwash_backend/internal/domain/modbus/repository"
	sessionModels "carwash_backend/internal/domain/session/models"
)

// ModbusService предоставляет методы для работы с Modbus устройствами
type ModbusService struct {
	db         *gorm.DB
	config     *config.Config
	repository *repository.ModbusRepository
}

// NewModbusService создает новый экземпляр ModbusService
func NewModbusService(db *gorm.DB, config *config.Config) *ModbusService {
	return &ModbusService{
		db:         db,
		config:     config,
		repository: repository.NewModbusRepository(db),
	}
}

// WriteCoil записывает значение в coil Modbus устройства
func (s *ModbusService) WriteCoil(boxID uuid.UUID, register string, value bool) error {
	operation := &models.ModbusOperation{
		ID:        uuid.New(),
		BoxID:     boxID,
		Operation: fmt.Sprintf("write_coil_%s", register),
		Register:  register,
		Value:     value,
		CreatedAt: time.Now(),
	}

	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		log.Printf("Modbus отключен, пропускаем запись в регистр %s - box_id: %s, register: %s, value: %v",
			register, boxID, register, value)

		// Сохраняем операцию как успешную (так как Modbus отключен намеренно)
		operation.Success = true
		operation.Error = "Modbus отключен в конфигурации"
		s.repository.SaveModbusOperation(operation)
		return nil
	}

	// Получаем конфигурацию бокса
	_, err := s.repository.GetWashBoxByID(boxID)
	if err != nil {
		operation.Success = false
		operation.Error = fmt.Sprintf("не удалось найти бокс: %v", err)
		s.repository.SaveModbusOperation(operation)
		log.Printf("Ошибка поиска бокса - box_id: %s, error: %v", boxID, err)
		return fmt.Errorf("не удалось найти бокс: %v", err)
	}

	// Проверяем, что регистр задан
	if register == "" {
		operation.Success = false
		operation.Error = fmt.Sprintf("регистр не задан для бокса %s", boxID)
		s.repository.SaveModbusOperation(operation)
		log.Printf("Регистр не задан - box_id: %s", boxID)
		return fmt.Errorf("регистр не задан для бокса %s", boxID)
	}

	// Проверяем формат регистра (hex)
	if !s.isValidHexRegister(register) {
		operation.Success = false
		operation.Error = fmt.Sprintf("неверный формат регистра: %s", register)
		s.repository.SaveModbusOperation(operation)
		log.Printf("Неверный формат регистра - box_id: %s, register: %s", boxID, register)
		return fmt.Errorf("неверный формат регистра: %s", register)
	}

	// Создаем Modbus клиент
	client, handler, err := s.createModbusClient()
	if err != nil {
		operation.Success = false
		operation.Error = fmt.Sprintf("не удалось создать Modbus клиент: %v", err)
		s.repository.SaveModbusOperation(operation)
		s.repository.UpdateModbusConnectionStatus(boxID, false, err.Error())
		log.Printf("Ошибка создания Modbus клиента - box_id: %s, error: %v", boxID, err)
		return fmt.Errorf("не удалось создать Modbus клиент: %v", err)
	}
	defer handler.Close()

	// Конвертируем hex регистр в uint16
	address, err := s.hexToUint16(register)
	if err != nil {
		operation.Success = false
		operation.Error = fmt.Sprintf("неверный формат регистра: %v", err)
		s.repository.SaveModbusOperation(operation)
		log.Printf("Ошибка конвертации регистра - box_id: %s, register: %s, error: %v", boxID, register, err)
		return fmt.Errorf("неверный формат регистра: %v", err)
	}

	// Выполняем запись с retry механизмом
	var lastError error
	for attempt := 1; attempt <= 3; attempt++ {
		// Записываем значение в coil
		_, err := client.WriteSingleCoil(address, uint16(boolToUint16(value)))
		if err == nil {
			operation.Success = true
			s.repository.SaveModbusOperation(operation)
			s.repository.UpdateModbusConnectionStatus(boxID, true, "")
			log.Printf("Modbus: успешно записано значение %v в регистр %s - box_id: %s, register: %s, value: %v",
				value, register, boxID, register, value)
			return nil
		}

		lastError = err
		log.Printf("Modbus: попытка %d неудачна - box_id: %s, register: %s, attempt: %d, error: %v",
			boxID, register, attempt, err)

		if attempt < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	operation.Success = false
	operation.Error = fmt.Sprintf("не удалось записать в Modbus после 3 попыток: %v", lastError)
	s.repository.SaveModbusOperation(operation)
	s.repository.UpdateModbusConnectionStatus(boxID, false, lastError.Error())
	log.Printf("Modbus: все попытки неудачны - box_id: %s, register: %s, error: %v", boxID, register, lastError)
	return fmt.Errorf("не удалось записать в Modbus после 3 попыток: %v", lastError)
}

// WriteLightCoil включает или выключает свет для бокса
func (s *ModbusService) WriteLightCoil(boxID uuid.UUID, value bool) error {
	washbox, err := s.repository.GetWashBoxByID(boxID)
	if err != nil {
		log.Printf("Ошибка поиска бокса для света - box_id: %s, error: %v", boxID, err)
		return fmt.Errorf("не удалось найти бокс: %v", err)
	}

	if washbox.LightCoilRegister == nil || *washbox.LightCoilRegister == "" {
		log.Printf("Регистр света не настроен - box_id: %s", boxID)
		return fmt.Errorf("регистр света не задан для бокса %s", boxID)
	}

	log.Printf("Управление светом - box_id: %s, register: %s, value: %v", boxID, *washbox.LightCoilRegister, value)
	return s.WriteCoil(boxID, *washbox.LightCoilRegister, value)
}

// WriteChemistryCoil включает или выключает химию для бокса
func (s *ModbusService) WriteChemistryCoil(boxID uuid.UUID, value bool) error {
	washbox, err := s.repository.GetWashBoxByID(boxID)
	if err != nil {
		log.Printf("Ошибка поиска бокса для химии - box_id: %s, error: %v", boxID, err)
		return fmt.Errorf("не удалось найти бокс: %v", err)
	}

	if washbox.ChemistryCoilRegister == nil || *washbox.ChemistryCoilRegister == "" {
		log.Printf("Регистр химии не настроен - box_id: %s", boxID)
		return fmt.Errorf("регистр химии не задан для бокса %s", boxID)
	}

	log.Printf("Управление химией - box_id: %s, register: %s, value: %v", boxID, *washbox.ChemistryCoilRegister, value)
	return s.WriteCoil(boxID, *washbox.ChemistryCoilRegister, value)
}

// HandleModbusError обрабатывает ошибку Modbus и продлевает время сессии
func (s *ModbusService) HandleModbusError(boxID uuid.UUID, operation string, sessionID uuid.UUID, err error) error {
	log.Printf("Modbus error handler - box_id: %s, operation: %s, session_id: %s, error: %v",
		boxID, operation, sessionID, err)

	// Проверяем, не связана ли ошибка с отключенным Modbus
	if err != nil && (err.Error() == "Modbus протокол отключен в конфигурации" ||
		err.Error() == "Modbus отключен, пропускаем запись в регистр") {
		// Если Modbus отключен, не продлеваем время сессии
		log.Printf("Modbus отключен, не продлеваем время сессии - session_id: %s", sessionID)
		return err
	}

	// Продлеваем время сессии на 5 минут только для реальных ошибок Modbus
	if err := s.extendSessionTime(sessionID, 5*time.Minute); err != nil {
		log.Printf("Не удалось продлить время сессии - session_id: %s, error: %v", sessionID, err)
	}

	// Обновляем статус подключения
	if err != nil {
		s.repository.UpdateModbusConnectionStatus(boxID, false, err.Error())
	}

	return err
}

// TestConnection тестирует соединение с Modbus устройством
func (s *ModbusService) TestConnection(boxID uuid.UUID) (*models.TestModbusConnectionResponse, error) {
	log.Printf("Тест подключения Modbus - box_id: %s", boxID)

	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		log.Printf("Modbus отключен в конфигурации - box_id: %s", boxID)
		return &models.TestModbusConnectionResponse{
			Success: false,
			Message: "Modbus протокол отключен в конфигурации",
		}, nil
	}

	// Проверяем, что бокс существует
	_, err := s.repository.GetWashBoxByID(boxID)
	if err != nil {
		log.Printf("Бокс не найден для теста подключения - box_id: %s, error: %v", boxID, err)
		return nil, fmt.Errorf("не удалось найти бокс: %v", err)
	}

	// Создаем Modbus клиент
	client, handler, err := s.createModbusClient()
	if err != nil {
		log.Printf("Ошибка создания Modbus клиента для теста - box_id: %s, error: %v", boxID, err)
		s.repository.UpdateModbusConnectionStatus(boxID, false, err.Error())
		return &models.TestModbusConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Не удалось создать Modbus клиент: %v", err),
		}, nil
	}
	defer handler.Close()

	// Пытаемся прочитать регистр для проверки соединения
	_, err = client.ReadCoils(1, 1)
	if err != nil {
		log.Printf("Ошибка чтения coil для теста - box_id: %s, error: %v", boxID, err)
		s.repository.UpdateModbusConnectionStatus(boxID, false, err.Error())
		return &models.TestModbusConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Не удалось подключиться к Modbus устройству: %v", err),
		}, nil
	}

	log.Printf("Тест подключения Modbus успешен - box_id: %s", boxID)
	s.repository.UpdateModbusConnectionStatus(boxID, true, "")
	return &models.TestModbusConnectionResponse{
		Success: true,
		Message: "Соединение с Modbus устройством установлено успешно",
	}, nil
}

// TestCoil тестирует запись в конкретный регистр
func (s *ModbusService) TestCoil(boxID uuid.UUID, register string, value bool) (*models.TestModbusCoilResponse, error) {
	log.Printf("Тест записи coil - box_id: %s, register: %s, value: %v", boxID, register, value)

	// Проверяем, включен ли Modbus
	if !s.config.ModbusEnabled {
		log.Printf("Modbus отключен для теста coil - box_id: %s, register: %s", boxID, register)
		return &models.TestModbusCoilResponse{
			Success: false,
			Message: "Modbus протокол отключен в конфигурации",
		}, nil
	}

	// Проверяем формат регистра
	if !s.isValidHexRegister(register) {
		log.Printf("Неверный формат регистра для теста - box_id: %s, register: %s", boxID, register)
		return &models.TestModbusCoilResponse{
			Success: false,
			Message: fmt.Sprintf("Неверный формат регистра: %s", register),
		}, nil
	}

	// Тестируем запись
	err := s.WriteCoil(boxID, register, value)
	if err != nil {
		log.Printf("Ошибка теста записи coil - box_id: %s, register: %s, value: %v, error: %v",
			boxID, register, value, err)
		return &models.TestModbusCoilResponse{
			Success: false,
			Message: fmt.Sprintf("Ошибка записи в регистр: %v", err),
		}, nil
	}

	log.Printf("Тест записи coil успешен - box_id: %s, register: %s, value: %v", boxID, register, value)
	return &models.TestModbusCoilResponse{
		Success: true,
		Message: fmt.Sprintf("Успешно записано значение %v в регистр %s", value, register),
	}, nil
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

	// Для локального тестирования используем HTTP API вместо прямого Modbus TCP
	if s.config.ModbusHost == "localhost" || s.config.ModbusHost == "127.0.0.1" {
		return nil, nil, fmt.Errorf("Локальное тестирование через HTTP API (порт 5001)")
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

// extendSessionTime продлевает время сессии
func (s *ModbusService) extendSessionTime(sessionID uuid.UUID, duration time.Duration) error {
	var session sessionModels.Session
	if err := s.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return fmt.Errorf("не удалось найти сессию: %v", err)
	}

	// Продлеваем время в зависимости от статуса сессии
	if session.Status == "in_queue" {
		// Для сессий в очереди продлеваем created_at
		newCreatedAt := session.CreatedAt.Add(-duration)
		return s.db.Model(&session).Update("created_at", newCreatedAt).Error
	} else if session.Status == "active" {
		// Для активных сессий продлеваем status_updated_at
		newStatusUpdatedAt := session.StatusUpdatedAt.Add(-duration)
		return s.db.Model(&session).Update("status_updated_at", newStatusUpdatedAt).Error
	}

	return nil
}

// GetStatus получает статус Modbus устройства для бокса
func (s *ModbusService) GetStatus(boxID uuid.UUID) (*models.GetModbusStatusResponse, error) {
	log.Printf("Получение статуса Modbus - box_id: %s", boxID)

	// Получаем информацию о боксе
	_, err := s.repository.GetWashBoxByID(boxID)
	if err != nil {
		log.Printf("Ошибка поиска бокса для статуса - box_id: %s, error: %v", boxID, err)
		return nil, fmt.Errorf("не удалось найти бокс: %v", err)
	}

	// Получаем статус подключения
	connectionStatus, err := s.repository.GetModbusConnectionStatus(boxID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка получения статуса подключения - box_id: %s, error: %v", boxID, err)
		return nil, fmt.Errorf("ошибка получения статуса подключения: %v", err)
	}

	// Формируем ответ
	status := models.ModbusStatus{
		BoxID: boxID,
		Host:  s.config.ModbusHost,
		Port:  s.config.ModbusPort,
	}

	if connectionStatus != nil {
		status.Connected = connectionStatus.Connected
		status.LastError = connectionStatus.LastError
		status.LastSeen = connectionStatus.LastSeen
	} else {
		status.Connected = false
		status.LastError = "Статус подключения не определен"
	}

	log.Printf("Статус Modbus получен - box_id: %s, connected: %v", boxID, status.Connected)
	return &models.GetModbusStatusResponse{
		Status: status,
	}, nil
}

// GetDashboard получает данные для дашборда мониторинга
func (s *ModbusService) GetDashboard(timeRange string) (*models.GetModbusDashboardResponse, error) {
	log.Printf("Получение данных дашборда Modbus - time_range: %s", timeRange)

	// Определяем временной диапазон
	var since time.Time
	switch timeRange {
	case "1h":
		since = time.Now().Add(-time.Hour)
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		since = time.Now().Add(-30 * 24 * time.Hour)
	default:
		since = time.Now().Add(-24 * time.Hour) // По умолчанию 24 часа
	}

	// Получаем боксы с настроенным modbus
	boxes, err := s.repository.GetActiveBoxesWithModbusConfig()
	if err != nil {
		log.Printf("Ошибка получения боксов для дашборда: %v", err)
		return nil, fmt.Errorf("ошибка получения боксов: %v", err)
	}

	// Получаем статусы подключений
	connectionStatuses, err := s.repository.GetAllModbusConnectionStatuses()
	if err != nil {
		log.Printf("Ошибка получения статусов подключений: %v", err)
		return nil, fmt.Errorf("ошибка получения статусов подключений: %v", err)
	}

	// Создаем мапу статусов для быстрого доступа
	statusMap := make(map[uuid.UUID]*models.ModbusConnectionStatus)
	for i := range connectionStatuses {
		statusMap[connectionStatuses[i].BoxID] = &connectionStatuses[i]
	}

	// Формируем обзор
	overview := models.ModbusDashboardOverview{
		TotalBoxes: len(boxes),
	}

	// Формируем статусы боксов
	var boxStatuses []models.ModbusBoxStatus
	for _, box := range boxes {
		boxStatus := models.ModbusBoxStatus{
			BoxID:                 box.ID,
			BoxNumber:             box.Number,
			LightCoilRegister:     box.LightCoilRegister,
			ChemistryCoilRegister: box.ChemistryCoilRegister,
		}

		// Добавляем информацию о подключении
		if status, exists := statusMap[box.ID]; exists {
			boxStatus.Connected = status.Connected
			boxStatus.LastError = status.LastError
			boxStatus.LastSeen = status.LastSeen
			if status.Connected {
				overview.ConnectedBoxes++
			} else {
				overview.DisconnectedBoxes++
			}
		} else {
			boxStatus.Connected = false
			boxStatus.LastError = "Статус не определен"
			overview.DisconnectedBoxes++
		}

		// Получаем статистику операций для бокса
		stats, err := s.repository.GetModbusStats(&box.ID, since)
		if err == nil {
			boxStatus.OperationsLast24h = stats.TotalOperations
			if stats.TotalOperations > 0 {
				boxStatus.SuccessRateLast24h = float64(stats.SuccessfulOperations) / float64(stats.TotalOperations) * 100
			}
		}

		boxStatuses = append(boxStatuses, boxStatus)
	}

	// Получаем общую статистику операций
	totalStats, err := s.repository.GetModbusStats(nil, since)
	if err == nil {
		overview.OperationsLast24h = totalStats.TotalOperations
		if totalStats.TotalOperations > 0 {
			overview.SuccessRateLast24h = float64(totalStats.SuccessfulOperations) / float64(totalStats.TotalOperations) * 100
		}
	}

	// Получаем последние операции
	recentOperations, err := s.repository.GetModbusOperations(nil, 20, 0)
	if err != nil {
		log.Printf("Ошибка получения последних операций: %v", err)
		recentOperations = []models.ModbusOperation{}
	}

	// Получаем статистику ошибок
	errorStats, err := s.repository.GetModbusErrorsByType(nil, since)
	if err != nil {
		log.Printf("Ошибка получения статистики ошибок: %v", err)
		errorStats = make(map[string]int64)
	}

	data := models.ModbusDashboardData{
		Overview:         overview,
		BoxStatuses:      boxStatuses,
		RecentOperations: recentOperations,
		ErrorStats:       errorStats,
	}

	log.Printf("Данные дашборда Modbus сформированы - boxes: %d, operations: %d",
		len(boxes), overview.OperationsLast24h)

	return &models.GetModbusDashboardResponse{
		Data: data,
	}, nil
}

// GetHistory получает историю операций Modbus
func (s *ModbusService) GetHistory(req *models.GetModbusHistoryRequest) (*models.GetModbusHistoryResponse, error) {
	log.Printf("Получение истории Modbus - box_id: %v, limit: %d, offset: %d",
		req.BoxID, req.Limit, req.Offset)

	// Устанавливаем значения по умолчанию
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Получаем операции
	operations, err := s.repository.GetModbusOperations(req.BoxID, req.Limit, req.Offset)
	if err != nil {
		log.Printf("Ошибка получения истории операций: %v", err)
		return nil, fmt.Errorf("ошибка получения истории операций: %v", err)
	}

	// TODO: Добавить фильтрацию по операции, успешности и времени
	// Пока возвращаем базовый результат

	response := &models.GetModbusHistoryResponse{
		Operations: operations,
		Total:      int64(len(operations)), // TODO: получать реальный count
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	log.Printf("История Modbus получена - операций: %d", len(operations))
	return response, nil
}
