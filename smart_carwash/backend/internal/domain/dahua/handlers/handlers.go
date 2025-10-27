package handlers

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"carwash_backend/internal/domain/dahua/models"
	"carwash_backend/internal/domain/dahua/service"
)

// Handler представляет HTTP обработчики для Dahua интеграции
type Handler struct {
	dahuaService service.Service
}

// NewHandler создает новый экземпляр обработчика
func NewHandler(dahuaService service.Service) *Handler {
	return &Handler{
		dahuaService: dahuaService,
	}
}

// ANPRWebhook обрабатывает webhook от камеры Dahua в формате ITSAPI XML
// POST /api/v1/dahua/anpr-webhook
func (h *Handler) ANPRWebhook(c *gin.Context) {
	// Читаем тело запроса для логирования
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Ошибка чтения запроса: %v", err)
		c.Header("Content-Type", "application/xml")
		c.String(http.StatusBadRequest, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <result>ERROR</result>
    <message>Ошибка чтения запроса</message>
</Response>`)
		return
	}

	// Восстанавливаем body для повторного чтения
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	// Определяем Content-Type для выбора формата парсинга
	contentType := c.GetHeader("Content-Type")

	var webhookReq *models.DahuaWebhookRequest
	var processReq *models.ProcessANPREventRequest

	// Парсим входящие данные в зависимости от Content-Type
	if contentType == "application/xml" || contentType == "text/xml" {
		// Парсинг XML формата ITSAPI
		var xmlReq models.DahuaWebhookRequest
		if err := c.ShouldBindXML(&xmlReq); err != nil {
			log.Printf("❌ Ошибка парсинга XML: %v", err)
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusBadRequest, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <result>ERROR</result>
    <message>Неверный формат XML: %s</message>
</Response>`, err.Error())
			return
		}
		webhookReq = &xmlReq
		processReq = xmlReq.ToProcessRequest()
	} else {
		// Парсинг JSON формата (реальная структура от камеры Dahua)
		var jsonReq models.DahuaWebhookRequestJSON
		if err := c.ShouldBindJSON(&jsonReq); err != nil {
			log.Printf("❌ Ошибка парсинга JSON: %v", err)
			c.JSON(http.StatusBadRequest, models.DahuaWebhookResponseJSON{
				Success: false,
				Message: "Неверный формат JSON: " + err.Error(),
			})
			return
		}

		// Проверяем, что номер автомобиля существует
		if !jsonReq.ValidatePlateNumber() {
			log.Printf("❌ Номер автомобиля не найден в данных")
			c.JSON(http.StatusBadRequest, models.DahuaWebhookResponseJSON{
				Success: false,
				Message: "Номер автомобиля не найден",
			})
			return
		}

		// Преобразуем JSON в XML структуру для единообразной обработки
		webhookReq = &models.DahuaWebhookRequest{
			LicensePlate: jsonReq.GetPlateNumber(),
			Confidence:   100,   // Считаем что камера уверена
			Direction:    "out", // Любой запрос означает выезд
			EventType:    "ANPR",
			DateTime:     time.Now().Format("2006-01-02T15:04:05"),
			ImagePath:    "",
		}
		processReq = jsonReq.ToProcessRequest()
	}

	// Валидируем направление
	if !webhookReq.ValidateDirection() {
		if contentType == "application/xml" || contentType == "text/xml" {
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusBadRequest, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <result>ERROR</result>
    <message>Неверное направление движения. Допустимые значения: in, out</message>
</Response>`)
		} else {
			c.JSON(http.StatusBadRequest, models.DahuaWebhookResponseJSON{
				Success: false,
				Message: "Неверное направление движения. Допустимые значения: in, out",
			})
		}
		return
	}

	// Обрабатываем событие
	response, err := h.dahuaService.ProcessANPREvent(processReq)
	if err != nil {
		log.Printf("❌ Ошибка обработки ANPR события: %v", err)
		if contentType == "application/xml" || contentType == "text/xml" {
			c.Header("Content-Type", "application/xml")
			c.String(http.StatusInternalServerError, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <result>ERROR</result>
    <message>%s</message>
</Response>`, err.Error())
		} else {
			c.JSON(http.StatusInternalServerError, models.DahuaWebhookResponseJSON{
				Success: false,
				Message: err.Error(),
			})
		}
		return
	}

	// Логируем результат обработки
	log.Printf("✅ ANPR событие обработано успешно: %s", response.Message)
	if response.UserFound {
		log.Printf("👤 Пользователь найден: %s", webhookReq.LicensePlate)
	}
	if response.SessionFound {
		log.Printf("🎯 Активная сессия найдена: %s", response.SessionID)
	}

	// Возвращаем ответ в соответствующем формате
	if contentType == "application/xml" || contentType == "text/xml" {
		c.Header("Content-Type", "application/xml")
		if response.Success {
			c.String(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <result>OK</result>
    <message>%s</message>
</Response>`, response.Message)
		} else {
			c.String(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <result>ERROR</result>
    <message>%s</message>
</Response>`, response.Message)
		}
	} else {
		c.JSON(http.StatusOK, models.DahuaWebhookResponseJSON{
			Success: response.Success,
			Message: response.Message,
		})
	}
}

// HealthCheck проверяет состояние Dahua интеграции
// GET /api/v1/dahua/health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dahua интеграция работает",
		"service": "dahua",
	})
}

// DeviceInfo обрабатывает запросы регистрации устройства от камеры Dahua
// POST /NotificationInfo/DeviceInfo
func (h *Handler) DeviceInfo(c *gin.Context) {
	// Читаем тело запроса для логирования
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Ошибка чтения запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"Result":  "Error",
			"Message": "Ошибка чтения запроса",
		})
		return
	}

	// Восстанавливаем body для повторного чтения
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	// Определяем Content-Type для выбора формата ответа
	contentType := c.GetHeader("Content-Type")

	// Логируем все параметры запроса
	log.Printf("📨 Запрос на /NotificationInfo/DeviceInfo")
	log.Printf("📋 Method: %s", c.Request.Method)
	log.Printf("📋 Headers: %v", c.Request.Header)
	log.Printf("📄 Body: %s", string(body))
	log.Printf("📋 Query params: %v", c.Request.URL.Query())
	log.Printf("📋 Content-Type: %s", contentType)
	log.Printf("📋 Client IP: %s", c.ClientIP())

	// Пытаемся распарсить как JSON (регистрация устройства)
	var deviceInfo models.DahuaDeviceRegistration
	if err := c.ShouldBindJSON(&deviceInfo); err != nil {
		log.Printf("❌ Ошибка парсинга JSON от камеры: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"Result":  "Error",
			"Message": "Invalid JSON format",
		})
		return
	}

	// Отправляем подтверждение регистрации
	response := models.DeviceRegistrationResponse{
		Result:    "OK",
		DeviceID:  deviceInfo.DeviceID,
		Message:   "Device registered successfully",
		Timestamp: time.Now().Format("2006-01-02T15:04:05+08:00"),
		ServerID:  "carwash-server-001",
		Status:    "Online",
	}

	log.Printf("✅ Камера успешно зарегистрирована: %s (%s)",
		deviceInfo.DeviceName, deviceInfo.IPAddress)

	c.Header("Content-Type", "application/json;charset=UTF-8")
	c.JSON(http.StatusOK, response)
}

// KeepAlive обрабатывает heartbeat запросы от камеры Dahua
// GET/POST /NotificationInfo/KeepAlive
func (h *Handler) KeepAlive(c *gin.Context) {
	// Читаем тело запроса для логирования (если это POST)
	var bodyContent string
	if c.Request.Method == "POST" {
		body, err := io.ReadAll(c.Request.Body)
		if err == nil {
			bodyContent = string(body)
			// Восстанавливаем body для повторного чтения
			c.Request.Body = io.NopCloser(strings.NewReader(bodyContent))
		}
	}

	// Логируем heartbeat от камеры
	log.Printf("💓 Heartbeat от камеры на /NotificationInfo/KeepAlive")
	log.Printf("📋 Method: %s", c.Request.Method)
	log.Printf("📋 Client IP: %s", c.ClientIP())
	if bodyContent != "" {
		log.Printf("💓 Heartbeat body: %s", bodyContent)
	}

	// Определяем Content-Type для выбора формата ответа
	contentType := c.GetHeader("Content-Type")

	// Возвращаем ответ в соответствующем формате
	if contentType == "application/xml" || contentType == "text/xml" {
		// XML ответ в формате ITSAPI
		c.Header("Content-Type", "application/xml; charset=UTF-8")
		c.String(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>
<HeartbeatResponse>
    <result>OK</result>
    <timestamp>%s</timestamp>
    <status>online</status>
    <nextHeartbeatInterval>30</nextHeartbeatInterval>
</HeartbeatResponse>`, time.Now().Format("2006-01-02T15:04:05+08:00"))
	} else {
		// JSON ответ (для обратной совместимости)
		c.JSON(http.StatusOK, gin.H{
			"success":               true,
			"message":               "Heartbeat received",
			"timestamp":             time.Now().Format("2006-01-02T15:04:05+08:00"),
			"status":                "online",
			"nextHeartbeatInterval": 30,
		})
	}
}
