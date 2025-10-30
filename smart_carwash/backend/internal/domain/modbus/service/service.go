package service

import (
	"carwash_backend/internal/logger"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/modbus/client"
	"carwash_backend/internal/domain/modbus/models"
	"carwash_backend/internal/domain/modbus/repository"
)

// ModbusService предоставляет методы для работы с Modbus через HTTP клиент и БД
type ModbusService struct {
	httpClient *client.ModbusHTTPClient
	repository *repository.ModbusRepository
	config     *config.Config
}

// NewModbusService создает новый экземпляр ModbusService
func NewModbusService(db *gorm.DB, config *config.Config) *ModbusService {
	return &ModbusService{
		httpClient: client.NewModbusHTTPClient(config),
		repository: repository.NewModbusRepository(db),
		config:     config,
	}
}

// TestCoil тестирует запись в конкретный регистр через HTTP клиент
func (s *ModbusService) TestCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) (*models.TestModbusCoilResponse, error) {
	logger.Printf("Тест записи coil через HTTP клиент - box_id: %s, register: %s, value: %v", boxID, register, value)

	// Проверяем, что бокс существует
	box, err := s.repository.GetWashBoxByID(ctx, boxID)
	if err != nil {
		logger.Printf("Бокс не найден для теста coil - box_id: %s, error: %v", boxID, err)
		return nil, fmt.Errorf("не удалось найти бокс: %v", err)
	}

	// Используем HTTP клиент для тестирования
	resp, err := s.httpClient.TestCoil(ctx, boxID, register, value)

	// Определяем тип операции по регистру
	operation := "test_coil"
	coilType := ""
	if box.LightCoilRegister != nil && *box.LightCoilRegister == register {
		operation = "test_light"
		coilType = "light"
	} else if box.ChemistryCoilRegister != nil && *box.ChemistryCoilRegister == register {
		operation = "test_chemistry"
		coilType = "chemistry"
	}

	// Сохраняем операцию в БД
	modbusOp := &models.ModbusOperation{
		ID:        uuid.New(),
		BoxID:     boxID,
		Operation: operation,
		Register:  register,
		Value:     value,
		Success:   err == nil && resp != nil && resp.Success,
		CreatedAt: time.Now(),
	}

	if err != nil {
		modbusOp.Error = err.Error()
		logger.Printf("Ошибка HTTP клиента для теста coil - box_id: %s, error: %v", boxID, err)

		// Сохраняем операцию с ошибкой
		if saveErr := s.repository.SaveModbusOperation(ctx, modbusOp); saveErr != nil {
			logger.Printf("Ошибка сохранения операции TestCoil - box_id: %s, error: %v", boxID, saveErr)
		}

		return &models.TestModbusCoilResponse{
			Success: false,
			Message: fmt.Sprintf("Ошибка HTTP клиента: %v", err),
		}, nil
	}

	if !resp.Success {
		modbusOp.Error = resp.Message
	}

	// Сохраняем операцию
	if saveErr := s.repository.SaveModbusOperation(ctx, modbusOp); saveErr != nil {
		logger.Printf("Ошибка сохранения операции TestCoil - box_id: %s, error: %v", boxID, saveErr)
	}

	// Если операция успешна и это известный койл, обновляем его статус
	if resp.Success && coilType != "" {
		if updateErr := s.repository.UpdateModbusCoilStatus(ctx, boxID, coilType, value); updateErr != nil {
			logger.Printf("Ошибка обновления статуса койла при тесте - box_id: %s, coil_type: %s, error: %v", boxID, coilType, updateErr)
		}
	}

	logger.Printf("Тест записи coil через HTTP клиент завершен - box_id: %s, success: %v", boxID, resp.Success)
	return &models.TestModbusCoilResponse{
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

// GetStatus получает статус Modbus устройства для бокса из БД
func (s *ModbusService) GetStatus(ctx context.Context, boxID uuid.UUID) (*models.GetModbusStatusResponse, error) {
	logger.Printf("Получение статуса Modbus из БД - box_id: %s", boxID)

	// Получаем информацию о боксе
	_, err := s.repository.GetWashBoxByID(ctx, boxID)
	if err != nil {
		logger.Printf("Ошибка поиска бокса для статуса - box_id: %s, error: %v", boxID, err)
		return nil, fmt.Errorf("не удалось найти бокс: %v", err)
	}

	// Получаем статус подключения из БД
	connectionStatus, err := s.repository.GetModbusConnectionStatus(ctx, boxID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Printf("Ошибка получения статуса подключения - box_id: %s, error: %v", boxID, err)
		return nil, fmt.Errorf("ошибка получения статуса подключения: %v", err)
	}

	// Формируем ответ
	status := models.ModbusStatus{
		BoxID: boxID,
		Host:  s.config.ModbusServerHost,
		Port:  s.config.ModbusServerPort,
	}

	if connectionStatus != nil {
		status.Connected = connectionStatus.Connected
		status.LastError = connectionStatus.LastError
		status.LastSeen = connectionStatus.LastSeen
	} else {
		status.Connected = false
		status.LastError = "Статус подключения не определен"
	}

	logger.Printf("Статус Modbus получен из БД - box_id: %s, connected: %v", boxID, status.Connected)
	return &models.GetModbusStatusResponse{
		Status: status,
	}, nil
}

// GetDashboard получает данные для дашборда мониторинга из БД
func (s *ModbusService) GetDashboard(ctx context.Context, timeRange string) (*models.GetModbusDashboardResponse, error) {
	logger.Printf("Получение данных дашборда Modbus из БД - time_range: %s", timeRange)

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
	boxes, err := s.repository.GetActiveBoxesWithModbusConfig(ctx)
	if err != nil {
		logger.Printf("Ошибка получения боксов для дашборда: %v", err)
		return nil, fmt.Errorf("ошибка получения боксов: %v", err)
	}

	// Получаем статусы подключений
	connectionStatuses, err := s.repository.GetAllModbusConnectionStatuses(ctx)
	if err != nil {
		logger.Printf("Ошибка получения статусов подключений: %v", err)
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

		// Добавляем информацию о подключении и статусы койлов
		if status, exists := statusMap[box.ID]; exists {
			boxStatus.Connected = status.Connected
			boxStatus.LastError = status.LastError
			boxStatus.LightStatus = status.LightStatus
			boxStatus.ChemistryStatus = status.ChemistryStatus

			// Бокс считается подключенным, если у него включен свет
			if status.LightStatus != nil && *status.LightStatus {
				overview.ConnectedBoxes++
			} else {
				overview.DisconnectedBoxes++
			}
		} else {
			boxStatus.Connected = false
			boxStatus.LastError = "Статус не определен"
			overview.DisconnectedBoxes++
		}

		boxStatuses = append(boxStatuses, boxStatus)
	}

	// Получаем общую статистику операций
	totalStats, err := s.repository.GetModbusStats(ctx, nil, since)
	if err == nil {
		overview.OperationsLast24h = totalStats.TotalOperations
		if totalStats.TotalOperations > 0 {
			overview.SuccessRateLast24h = float64(totalStats.SuccessfulOperations) / float64(totalStats.TotalOperations) * 100
		}
	}

	// Получаем последние операции
	recentOperations, err := s.repository.GetModbusOperations(ctx, nil, 20, 0)
	if err != nil {
		logger.Printf("Ошибка получения последних операций: %v", err)
		recentOperations = []models.ModbusOperation{}
	}

	// Получаем статистику ошибок
	errorStats, err := s.repository.GetModbusErrorsByType(ctx, nil, since)
	if err != nil {
		logger.Printf("Ошибка получения статистики ошибок: %v", err)
		errorStats = make(map[string]int64)
	}

	data := models.ModbusDashboardData{
		Overview:         overview,
		BoxStatuses:      boxStatuses,
		RecentOperations: recentOperations,
		ErrorStats:       errorStats,
	}

	logger.Printf("Данные дашборда Modbus сформированы из БД - boxes: %d, operations: %d",
		len(boxes), overview.OperationsLast24h)

	return &models.GetModbusDashboardResponse{
		Data: data,
	}, nil
}

// GetHistory получает историю операций Modbus из БД
func (s *ModbusService) GetHistory(ctx context.Context, req *models.GetModbusHistoryRequest) (*models.GetModbusHistoryResponse, error) {
	logger.Printf("Получение истории Modbus из БД - box_id: %v, limit: %d, offset: %d",
		req.BoxID, req.Limit, req.Offset)

	// Устанавливаем значения по умолчанию
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Получаем операции из БД
	operations, err := s.repository.GetModbusOperations(ctx, req.BoxID, req.Limit, req.Offset)
	if err != nil {
		logger.Printf("Ошибка получения истории операций: %v", err)
		return nil, fmt.Errorf("ошибка получения истории операций: %v", err)
	}

	// Получаем общее количество операций для пагинации
	totalCount, err := s.repository.GetModbusOperationsCount(ctx, req.BoxID)
	if err != nil {
		logger.Printf("Ошибка получения общего количества операций: %v", err)
		// Используем количество полученных операций как fallback
		totalCount = int64(len(operations))
	}

	// Применяем фильтрацию по операции, успешности и времени (если указаны)
	filteredOperations := operations
	if req.Operation != nil && *req.Operation != "" {
		var filtered []models.ModbusOperation
		for _, op := range operations {
			if op.Operation == *req.Operation {
				filtered = append(filtered, op)
			}
		}
		filteredOperations = filtered
	}

	if req.Success != nil {
		var filtered []models.ModbusOperation
		for _, op := range filteredOperations {
			if op.Success == *req.Success {
				filtered = append(filtered, op)
			}
		}
		filteredOperations = filtered
	}

	response := &models.GetModbusHistoryResponse{
		Operations: filteredOperations,
		Total:      totalCount,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	logger.Printf("История Modbus получена из БД - операций: %d", len(operations))
	return response, nil
}
