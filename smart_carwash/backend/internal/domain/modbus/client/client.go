package client

import (
	"bytes"
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
			Timeout: 30 * time.Second,
		},
	}
}

// WriteCoil записывает значение в coil Modbus устройства
func (c *ModbusHTTPClient) WriteCoil(boxID uuid.UUID, register string, value bool) error {
	req := WriteCoilRequest{
		BoxID:    boxID,
		Register: register,
		Value:    value,
	}

	var resp WriteCoilResponse
	err := c.makeRequest("POST", "/api/v1/modbus/coil", req, &resp)
	if err != nil {
		return fmt.Errorf("ошибка HTTP запроса: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("ошибка Modbus: %s", resp.Message)
	}

	return nil
}

// WriteLightCoil включает или выключает свет для бокса
func (c *ModbusHTTPClient) WriteLightCoil(boxID uuid.UUID, value bool) error {
	req := WriteLightCoilRequest{
		BoxID: boxID,
		Value: value,
	}

	var resp WriteLightCoilResponse
	err := c.makeRequest("POST", "/api/v1/modbus/light", req, &resp)
	if err != nil {
		return fmt.Errorf("ошибка HTTP запроса: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("ошибка Modbus: %s", resp.Message)
	}

	return nil
}

// WriteChemistryCoil включает или выключает химию для бокса
func (c *ModbusHTTPClient) WriteChemistryCoil(boxID uuid.UUID, value bool) error {
	req := WriteChemistryCoilRequest{
		BoxID: boxID,
		Value: value,
	}

	var resp WriteChemistryCoilResponse
	err := c.makeRequest("POST", "/api/v1/modbus/chemistry", req, &resp)
	if err != nil {
		return fmt.Errorf("ошибка HTTP запроса: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("ошибка Modbus: %s", resp.Message)
	}

	return nil
}

// TestCoil тестирует запись в конкретный регистр
func (c *ModbusHTTPClient) TestCoil(boxID uuid.UUID, register string, value bool) (*TestCoilResponse, error) {
	req := TestCoilRequest{
		BoxID:    boxID,
		Register: register,
		Value:    value,
	}

	var resp TestCoilResponse
	err := c.makeRequest("POST", "/api/v1/modbus/test-coil", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %v", err)
	}

	return &resp, nil
}

// makeRequest выполняет HTTP запрос
func (c *ModbusHTTPClient) makeRequest(method, path string, requestBody interface{}, responseBody interface{}) error {
	// Сериализуем тело запроса
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("ошибка сериализации запроса: %v", err)
	}

	// Создаем HTTP запрос
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP запроса: %v", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP ошибка %d: %s", resp.StatusCode, string(body))
	}

	// Десериализуем ответ
	if responseBody != nil {
		err = json.Unmarshal(body, responseBody)
		if err != nil {
			return fmt.Errorf("ошибка десериализации ответа: %v", err)
		}
	}

	return nil
}
