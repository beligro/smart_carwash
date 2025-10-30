package handlers

import (
	"carwash_backend/internal/logger"
	"net/http"
	"strconv"

	"carwash_backend/internal/domain/user/models"
	"carwash_backend/internal/domain/user/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов пользователей
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для пользователей
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("", h.createUser)
		userRoutes.GET("/by-telegram-id", h.getUserByTelegramID) // telegram_id в query параметре
		userRoutes.PUT("/car-number", h.updateCarNumber)         // Обновление номера машины
		userRoutes.PUT("/email", h.updateEmail)                  // Обновление email
	}

	// Административные маршруты
	adminRoutes := router.Group("/admin/users")
	{
		adminRoutes.GET("", h.adminListUsers)
		adminRoutes.GET("/by-id", h.adminGetUser)
	}
}

// getUserByTelegramID обработчик для получения пользователя по telegram_id
func (h *Handler) getUserByTelegramID(c *gin.Context) {
	// Получаем telegram_id из query параметра
	telegramIDStr := c.Query("telegram_id")
	if telegramIDStr == "" {
		logger.WithContext(c).Errorf("API Error - getUserByTelegramID: не указан Telegram ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан Telegram ID"})
		return
	}

	// Преобразуем строку в int64
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getUserByTelegramID: некорректный Telegram ID '%s', error: %v", telegramIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный Telegram ID"})
		return
	}

	// Получаем пользователя по telegram_id
	user, err := h.service.GetUserByTelegramID(c.Request.Context(), telegramID)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getUserByTelegramID: пользователь не найден для telegram_id: %d, error: %v", telegramID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден. Возможно, вы не зарегистрированы в боте. Пожалуйста, нажмите /start в боте."})
		return
	}

	// Возвращаем пользователя
	c.JSON(http.StatusOK, models.CreateUserResponse{User: *user})
}

// createUser обработчик для создания пользователя
func (h *Handler) createUser(c *gin.Context) {
	var req models.CreateUserRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("API Error - createUser: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем пользователя
	user, err := h.service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - createUser: ошибка создания пользователя, telegram_id: %d, error: %v", req.TelegramID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем созданного пользователя
	c.JSON(http.StatusOK, models.CreateUserResponse{User: *user})
}

// updateCarNumber обработчик для обновления номера машины
func (h *Handler) updateCarNumber(c *gin.Context) {
	var req models.UpdateCarNumberRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("API Error - updateCarNumber: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"user_id":    req.UserID.String(),
		"car_number": req.CarNumber,
	})

	// Обновляем номер машины
	resp, err := h.service.UpdateCarNumber(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - updateCarNumber: ошибка обновления номера машины, user_id: %s, car_number: %s, error: %v", req.UserID.String(), req.CarNumber, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, resp)
}

// updateEmail обработчик для обновления email
func (h *Handler) updateEmail(c *gin.Context) {
	var req models.UpdateEmailRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("API Error - updateEmail: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"user_id": req.UserID.String(),
		"email":   req.Email,
	})

	// Обновляем email
	resp, err := h.service.UpdateEmail(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - updateEmail: ошибка обновления email, user_id: %s, email: %s, error: %v", req.UserID.String(), req.Email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, resp)
}

// adminListUsers обработчик для получения списка пользователей для администратора
func (h *Handler) adminListUsers(c *gin.Context) {
	// Получаем параметры пагинации из query
	var req models.AdminListUsersRequest

	// Лимит
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = &limit
		}
	}

	// Смещение
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = &offset
		}
	}

	// Получаем список пользователей
	resp, err := h.service.AdminListUsers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"limit":  req.Limit,
		"offset": req.Offset,
		"total":  resp.Total,
	})

	c.JSON(http.StatusOK, resp)
}

// adminGetUser обработчик для получения пользователя по ID для администратора
func (h *Handler) adminGetUser(c *gin.Context) {
	var req models.AdminGetUserRequest

	// Получаем ID из query параметра
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID пользователя обязателен"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	req.ID = id

	// Получаем пользователя
	resp, err := h.service.AdminGetUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"user_id":     req.ID,
		"telegram_id": resp.User.TelegramID,
	})

	c.JSON(http.StatusOK, resp)
}
