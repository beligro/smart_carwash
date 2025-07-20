package tinkoff

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"carwash_backend/internal/domain/payment/models"
)

// Client HTTP клиент для работы с Tinkoff Kassa API
type Client struct {
	terminalKey string
	secretKey   string
	baseURL     string
	httpClient  *http.Client
}

// NewClient создает новый клиент Tinkoff API
func NewClient(terminalKey, secretKey, baseURL string) *Client {
	return &Client{
		terminalKey: terminalKey,
		secretKey:   secretKey,
		baseURL:     baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Init инициализирует платеж в Tinkoff
func (c *Client) Init(req *models.InitRequest) (*models.InitResponse, error) {
	log.Println("Init request: ", req)
	req.TerminalKey = c.terminalKey

	signature := c.generateToken(req.ToMap())
	req.Token = signature

	resp, err := c.makeRequest("POST", "/v2/Init", req)
	if err != nil {
		log.Println("error: ", err)
		return nil, err
	}

	var initResp models.InitResponse
	if err := json.Unmarshal(resp, &initResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа Init: %w", err)
	}

	if !initResp.Success {
		log.Println("initResp: ", initResp)
		return nil, &models.TinkoffError{
			ErrorCode: initResp.ErrorCode,
			Message:   initResp.Message,
		}
	}

	return &initResp, nil
}

// GetState получает статус платежа
func (c *Client) GetState(req *models.GetStateRequest) (*models.GetStateResponse, error) {
	req.TerminalKey = c.terminalKey

	// Создаем данные для подписи
	data := map[string]string{
		"TerminalKey": c.terminalKey,
		"PaymentId":   req.PaymentId,
	}
	signature := c.generateSignature(data)
	data["Token"] = signature

	resp, err := c.makeRequest("POST", "/v2/GetState", data)
	if err != nil {
		return nil, err
	}

	var stateResp models.GetStateResponse
	if err := json.Unmarshal(resp, &stateResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа GetState: %w", err)
	}

	if !stateResp.Success {
		return nil, &models.TinkoffError{
			ErrorCode: stateResp.ErrorCode,
			Message:   stateResp.Message,
		}
	}

	return &stateResp, nil
}

// Cancel отменяет платеж
func (c *Client) Cancel(req *models.CancelRequest) (*models.CancelResponse, error) {
	req.TerminalKey = c.terminalKey

	// Создаем данные для подписи
	data := map[string]string{
		"TerminalKey": c.terminalKey,
		"PaymentId":   req.PaymentId,
	}
	if req.Amount > 0 {
		data["Amount"] = fmt.Sprintf("%d", req.Amount)
	}
	signature := c.generateSignature(data)
	data["Token"] = signature

	resp, err := c.makeRequest("POST", "/v2/Cancel", data)
	if err != nil {
		return nil, err
	}

	var cancelResp models.CancelResponse
	if err := json.Unmarshal(resp, &cancelResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа Cancel: %w", err)
	}

	if !cancelResp.Success {
		return nil, &models.TinkoffError{
			ErrorCode: cancelResp.ErrorCode,
			Message:   cancelResp.Message,
		}
	}

	return &cancelResp, nil
}

// Refund создает возврат средств
func (c *Client) Refund(req *models.RefundRequest) (*models.RefundResponse, error) {
	req.TerminalKey = c.terminalKey

	// Создаем данные для подписи
	data := map[string]string{
		"TerminalKey": c.terminalKey,
		"PaymentId":   req.PaymentId,
		"Amount":      fmt.Sprintf("%d", req.Amount),
	}
	signature := c.generateSignature(data)
	data["Token"] = signature

	resp, err := c.makeRequest("POST", "/v2/Refund", data)
	if err != nil {
		return nil, err
	}

	var refundResp models.RefundResponse
	if err := json.Unmarshal(resp, &refundResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа Refund: %w", err)
	}

	if !refundResp.Success {
		return nil, &models.TinkoffError{
			ErrorCode: refundResp.ErrorCode,
			Message:   refundResp.Message,
		}
	}

	return &refundResp, nil
}

// Confirm подтверждает платеж
func (c *Client) Confirm(req *models.ConfirmRequest) (*models.ConfirmResponse, error) {
	req.TerminalKey = c.terminalKey

	// Создаем данные для подписи
	data := map[string]string{
		"TerminalKey": c.terminalKey,
		"PaymentId":   req.PaymentId,
		"Amount":      fmt.Sprintf("%d", req.Amount),
	}
	signature := c.generateSignature(data)
	data["Token"] = signature

	resp, err := c.makeRequest("POST", "/v2/Confirm", data)
	if err != nil {
		return nil, err
	}

	var confirmResp models.ConfirmResponse
	if err := json.Unmarshal(resp, &confirmResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа Confirm: %w", err)
	}

	if !confirmResp.Success {
		return nil, &models.TinkoffError{
			ErrorCode: confirmResp.ErrorCode,
			Message:   confirmResp.Message,
		}
	}

	return &confirmResp, nil
}

// GetOperations получает операции за период
func (c *Client) GetOperations(req *models.GetOperationsRequest) (*models.GetOperationsResponse, error) {
	req.TerminalKey = c.terminalKey

	// Создаем данные для подписи
	data := map[string]string{
		"TerminalKey": c.terminalKey,
		"StartDate":   req.StartDate,
		"EndDate":     req.EndDate,
	}
	signature := c.generateSignature(data)
	data["Token"] = signature

	resp, err := c.makeRequest("POST", "/v2/GetOperations", data)
	if err != nil {
		return nil, err
	}

	var operationsResp models.GetOperationsResponse
	if err := json.Unmarshal(resp, &operationsResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа GetOperations: %w", err)
	}

	if !operationsResp.Success {
		return nil, &models.TinkoffError{
			ErrorCode: operationsResp.ErrorCode,
			Message:   operationsResp.Message,
		}
	}

	return &operationsResp, nil
}

// makeRequest выполняет HTTP запрос к Tinkoff API
func (c *Client) makeRequest(method, endpoint string, data interface{}) ([]byte, error) {
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("ошибка маршалинга данных: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP ошибка %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// generateSignature генерирует подпись для запроса
func (c *Client) generateSignature(data map[string]string) string {
	// Сортируем ключи
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Создаем строку для подписи
	var values []string
	for _, k := range keys {
		values = append(values, data[k])
	}
	signatureString := strings.Join(values, "")

	// Создаем HMAC-SHA256 подпись
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write([]byte(signatureString))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) generateToken(params map[string]interface{}) string {
    // Создаем копию параметров
    tokenParams := make(map[string]interface{})
    for k, v := range params {
        tokenParams[k] = v
    }

	log.Println("tokenParams: ", tokenParams)
	log.Println("c.secretKey: ", c.secretKey)
    
    // Добавляем пароль
    tokenParams["Password"] = c.secretKey
    
    // Удаляем Token если есть
    delete(tokenParams, "Token")
    
    // Получаем отсортированные ключи
    keys := make([]string, 0, len(tokenParams))
    for k := range tokenParams {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    // ОТЛАДКА: выводим что участвует в токене
    log.Printf("Token generation params (sorted):\n")
    for _, key := range keys {
        log.Printf("%s: %v\n", key, tokenParams[key])
    }
    
    // Конкатенируем значения
    var values []string
    for _, key := range keys {
        valueStr := fmt.Sprintf("%v", tokenParams[key])
        values = append(values, valueStr)
    }
    
    concatenated := strings.Join(values, "")
    log.Printf("Concatenated string: %s\n", concatenated)
    
    hash := sha256.Sum256([]byte(concatenated))
    token := fmt.Sprintf("%x", hash)
    
    log.Printf("Generated token: %s\n", token)
    return token
}


// VerifyWebhookSignature проверяет подпись webhook'а
func (c *Client) VerifyWebhookSignature(data map[string]string, signature string) bool {
	expectedSignature := c.generateSignature(data)
	return expectedSignature == signature
} 