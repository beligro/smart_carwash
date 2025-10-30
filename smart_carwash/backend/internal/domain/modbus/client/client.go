package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"carwash_backend/internal/config"
)

// ModbusHTTPClient HTTP клиент для взаимодействия с modbus сервером
type ModbusHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewModbusHTTPClient создает новый HTTP клиент для modbus сервера
func NewModbusHTTPClient(config *config.Config) *ModbusHTTPClient {
	return &ModbusHTTPClient{
		baseURL: fmt.Sprintf("http://%s:%d", config.ModbusServerHost, config.ModbusServerPort),
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // ✅ Уменьшил таймаут
		},
	}
}

// WriteCoil записывает значение в coil Modbus устройства
func (c *ModbusHTTPClient) WriteCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) error {
	req := WriteCoilRequest{
		BoxID:    boxID,
		Register: register,
		Value:    value,
	}

	var resp WriteCoilResponse
	if err := c.makeRequest(ctx, "POST", "/api/v1/modbus/coil", req, &resp); err != nil {
		return fmt.Errorf("ошибка HTTP запроса: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("ошибка Modbus: %s", resp.Message)
	}

	return nil
}

// WriteLightCoil включает или выключает свет для бокса
func (c *ModbusHTTPClient) WriteLightCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) error {
	req := WriteLightCoilRequest{
		BoxID:    boxID,
		Register: register,
		Value:    value,
	}

	var resp WriteLightCoilResponse
	if err := c.makeRequest(ctx, "POST", "/api/v1/modbus/light", req, &resp); err != nil {
		return fmt.Errorf("ошибка HTTP запроса: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("ошибка Modbus: %s", resp.Message)
	}

	return nil
}

// WriteChemistryCoil включает или выключает химию для бокса
func (c *ModbusHTTPClient) WriteChemistryCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) error {
	req := WriteChemistryCoilRequest{
		BoxID:    boxID,
		Register: register,
		Value:    value,
	}

	var resp WriteChemistryCoilResponse
	if err := c.makeRequest(ctx, "POST", "/api/v1/modbus/chemistry", req, &resp); err != nil {
		return fmt.Errorf("ошибка HTTP запроса: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("ошибка Modbus: %s", resp.Message)
	}

	return nil
}

// TestCoil тестирует запись в конкретный регистр
func (c *ModbusHTTPClient) TestCoil(ctx context.Context, boxID uuid.UUID, register string, value bool) (*TestCoilResponse, error) {
	req := TestCoilRequest{
		BoxID:    boxID,
		Register: register,
		Value:    value,
	}

	var resp TestCoilResponse
	if err := c.makeRequest(ctx, "POST", "/api/v1/modbus/test-coil", req, &resp); err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %w", err)
	}

	return &resp, nil
}

// makeRequest выполняет HTTP запрос с контекстом
func (c *ModbusHTTPClient) makeRequest(ctx context.Context, method, path string, requestBody interface{}, responseBody interface{}) error {
	// Сериализуем тело запроса
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	// ✅ Создаем HTTP запрос с контекстом
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP запроса: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// ✅ Проверяем тип ошибки
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("таймаут запроса к Modbus серверу: %w", err)
		}
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("запрос к Modbus серверу отменён: %w", err)
		}
		return fmt.Errorf("ошибка выполнения HTTP запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP ошибка %d: %s", resp.StatusCode, string(body))
	}

	// Десериализуем ответ
	if responseBody != nil {
		if err := json.Unmarshal(body, responseBody); err != nil {
			return fmt.Errorf("ошибка десериализации ответа: %w", err)
		}
	}

	return nil
}
