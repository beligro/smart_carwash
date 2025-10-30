package tinkoff

import (
	"bytes"
	"carwash_backend/internal/logger"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"carwash_backend/internal/domain/payment/service"
)

// Client реализация Tinkoff API клиента
type Client struct {
	terminalKey string
	secretKey   string
	successURL  string
	failURL     string
	baseURL     string
	httpClient  *http.Client
}

// NewClient создает новый экземпляр Tinkoff клиента
func NewClient(terminalKey, secretKey, successURL, failURL string) service.TinkoffClient {
	return &Client{
		terminalKey: terminalKey,
		secretKey:   secretKey,
		successURL:  successURL,
		failURL:     failURL,
		baseURL:     "https://securepay.tinkoff.ru/v2",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePayment создает платеж в Tinkoff
func (c *Client) CreatePayment(orderID string, amount int, description string, receipt map[string]interface{}) (*service.TinkoffPaymentResponse, error) {
	// Формируем параметры запроса
	params := map[string]interface{}{
		"TerminalKey": c.terminalKey,
		"Amount":      amount,
		"OrderId":     orderID,
		"Description": description,
		"SuccessURL":  c.successURL,
		"FailURL":     c.failURL,
	}

	// Добавляем подпись
	params["Token"] = c.generateToken(params)

	// Добавляем Receipt только если он не пустой
	if receipt != nil && len(receipt) > 0 {
		// Передаем Receipt как объект (согласно документации Tinkoff)
		params["Receipt"] = receipt
	}

	// Отправляем запрос
	resp, err := c.sendRequest("POST", "/Init", params)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}

	// Парсим ответ
	var tinkoffResp service.TinkoffPaymentResponse
	if err := json.Unmarshal(resp, &tinkoffResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return &tinkoffResp, nil
}

// RefundPayment возвращает деньги за платеж
func (c *Client) RefundPayment(paymentID string, amount int) (*service.TinkoffRefundResponse, error) {
	// Формируем параметры запроса
	params := map[string]interface{}{
		"TerminalKey": c.terminalKey,
		"PaymentId":   paymentID,
		"Amount":      amount,
	}

	// Добавляем подпись
	params["Token"] = c.generateToken(params)

	// Отправляем запрос
	resp, err := c.sendRequest("POST", "/Cancel", params)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса возврата: %w", err)
	}

	// Парсим ответ
	var tinkoffResp service.TinkoffRefundResponse
	if err := json.Unmarshal(resp, &tinkoffResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа возврата: %w", err)
	}

	return &tinkoffResp, nil
}

// VerifyWebhookSignature проверяет подпись webhook
func (c *Client) VerifyWebhookSignature(data []byte, signature string) bool {
	// Создаем HMAC подпись
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write(data)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Сравниваем подписи
	return strings.EqualFold(expectedSignature, signature)
}

// sendRequest отправляет HTTP запрос к Tinkoff API
func (c *Client) sendRequest(method, endpoint string, params map[string]interface{}) ([]byte, error) {
	// Формируем URL
	requestURL := c.baseURL + endpoint

	// Подготавливаем данные для отправки
	var requestBody io.Reader
	if method == "POST" {
		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("ошибка маршалинга JSON: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	// Создаем запрос
	req, err := http.NewRequest(method, requestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка HTTP: %d, тело: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) generateToken(params map[string]interface{}) string {
	// Создаем копию параметров
	tokenParams := make(map[string]interface{})
	for k, v := range params {
		tokenParams[k] = v
	}

	logger.Printf("tokenParams: %v", tokenParams)
	logger.Printf("c.secretKey: %s", c.secretKey)

	// Добавляем пароль
	tokenParams["Password"] = c.secretKey

	// Удаляем Token если есть
	delete(tokenParams, "Token")

	// Удаляем Receipt из формирования токена (согласно документации Tinkoff)
	delete(tokenParams, "Receipt")

	// Получаем отсортированные ключи
	keys := make([]string, 0, len(tokenParams))
	for k := range tokenParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// ОТЛАДКА: выводим что участвует в токене
	logger.Printf("Token generation params (sorted):\n")
	for _, key := range keys {
		logger.Printf("%s: %v\n", key, tokenParams[key])
	}

	// Конкатенируем значения
	var values []string
	for _, key := range keys {
		valueStr := fmt.Sprintf("%v", tokenParams[key])
		values = append(values, valueStr)
	}

	concatenated := strings.Join(values, "")
	logger.Printf("Concatenated string: %s\n", concatenated)

	hash := sha256.Sum256([]byte(concatenated))
	token := fmt.Sprintf("%x", hash)

	logger.Printf("Generated token: %s\n", token)
	return token
}

// buildReceipt формирует чек для фискализации
func (c *Client) buildReceipt(amount int, email string) map[string]interface{} {
	// Используем переданный email или фоллбек на хардкод
	receiptEmail := email
	if receiptEmail == "" {
		receiptEmail = "yndx-aagrom-ijakag@yandex.ru"
	}

	receipt := map[string]interface{}{
		"Email":    receiptEmail,
		"Taxation": "usn_income_outcome",
		"Items": []map[string]interface{}{
			{
				"Name":     "Услуга автомойки",
				"Price":    amount,
				"Quantity": 1.00,
				"Amount":   amount,
				"Tax":      "none",
			},
		},
	}

	return receipt
}
