package handlers

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"carwash_backend/internal/domain/dahua/models"
	"carwash_backend/internal/domain/dahua/service"
	"carwash_backend/internal/logger"
	"context"
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
		logger.WithFields(logrus.Fields{
			"handler": "dahua",
			"method":  "ANPRWebhook",
			"error":   err,
		}).Error("Ошибка чтения запроса")
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
			logger.WithFields(logrus.Fields{
				"handler":     "dahua",
				"method":      "ANPRWebhook",
				"format":      "XML",
				"content_type": contentType,
				"error":       err,
			}).Error("Ошибка парсинга XML")
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
			logger.WithFields(logrus.Fields{
				"handler":      "dahua",
				"method":       "ANPRWebhook",
				"format":       "JSON",
				"content_type": contentType,
				"error":        err,
			}).Error("Ошибка парсинга JSON")
			c.JSON(http.StatusBadRequest, models.DahuaWebhookResponseJSON{
				Success: false,
				Message: "Неверный формат JSON: " + err.Error(),
			})
			return
		}

		// Проверяем, что номер автомобиля существует
		if !jsonReq.ValidatePlateNumber() {
			logger.WithFields(logrus.Fields{
				"handler": "dahua",
				"method":  "ANPRWebhook",
				"format":  "JSON",
			}).Error("Номер автомобиля не найден в данных")
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

	// Обогащаем контекст признаками Dahua/ANPR
	ctx := c.Request.Context()
	if v, ok := c.Get("dahua_authenticated"); ok {
		ctx = context.WithValue(ctx, "dahua_authenticated", v)
	}
	if v, ok := c.Get("dahua_username"); ok {
		ctx = context.WithValue(ctx, "dahua_username", v)
	}
	// Обрабатываем событие
	response, err := h.dahuaService.ProcessANPREvent(ctx, processReq)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"handler":      "dahua",
			"method":       "ANPRWebhook",
			"license_plate": webhookReq.LicensePlate,
			"direction":     webhookReq.Direction,
			"error":        err,
		}).Error("Ошибка обработки ANPR события")
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
	logFields := logrus.Fields{
		"handler":       "dahua",
		"method":        "ANPRWebhook",
		"license_plate":  webhookReq.LicensePlate,
		"direction":      webhookReq.Direction,
		"success":       response.Success,
		"message":       response.Message,
	}
	if response.SessionFound {
		logFields["session_id"] = response.SessionID
	}
	logger.WithFields(logFields).Info("ANPR событие обработано")

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
		logger.WithFields(logrus.Fields{
			"handler": "dahua",
			"method":  "DeviceInfo",
			"error":   err,
		}).Error("Ошибка чтения запроса")
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
	logger.WithFields(logrus.Fields{
		"handler":      "dahua",
		"method":       "DeviceInfo",
		"http_method":  c.Request.Method,
		"headers":      c.Request.Header,
		"body":         string(body),
		"query_params": c.Request.URL.Query(),
		"content_type": contentType,
		"client_ip":    c.ClientIP(),
	}).Info("Запрос на регистрацию устройства")

	// Пытаемся распарсить как JSON (регистрация устройства)
	var deviceInfo models.DahuaDeviceRegistration
	if err := c.ShouldBindJSON(&deviceInfo); err != nil {
		logger.WithFields(logrus.Fields{
			"handler": "dahua",
			"method":  "DeviceInfo",
			"error":   err,
		}).Error("Ошибка парсинга JSON от камеры")
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

	logger.WithFields(logrus.Fields{
		"handler":     "dahua",
		"method":      "DeviceInfo",
		"device_id":   deviceInfo.DeviceID,
		"device_name": deviceInfo.DeviceName,
		"ip_address":  deviceInfo.IPAddress,
		"device_type": deviceInfo.DeviceType,
		"manufacturer": deviceInfo.Manufacturer,
	}).Info("Камера успешно зарегистрирована")

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
