package handlers

import (
	"net/http"
	"strings"

	"carwash_backend/internal/domain/auth/models"
	"carwash_backend/internal/domain/auth/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов авторизации
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для авторизации
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	authRoutes := router.Group("/auth")
	{
		// Маршруты для авторизации
		authRoutes.POST("/admin/login", h.loginAdmin)
		authRoutes.POST("/cashier/login", h.loginCashier)
		authRoutes.POST("/logout", h.authMiddleware(), h.logout)

		// Маршруты для управления кассирами (только для администратора)
		cashierRoutes := authRoutes.Group("/cashiers", h.adminMiddleware())
		{
			cashierRoutes.POST("", h.createCashier)
			cashierRoutes.GET("", h.getCashiers)
			cashierRoutes.GET("/:id", h.getCashierByID)
			cashierRoutes.PUT("/:id", h.updateCashier)
			cashierRoutes.DELETE("/:id", h.deleteCashier)
		}
	}
}

// loginAdmin обработчик для авторизации администратора
func (h *Handler) loginAdmin(c *gin.Context) {
	var req models.LoginRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Авторизуем администратора
	resp, err := h.service.LoginAdmin(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем токен
	c.JSON(http.StatusOK, resp)
}

// loginCashier обработчик для авторизации кассира
func (h *Handler) loginCashier(c *gin.Context) {
	var req models.LoginRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Авторизуем кассира
	resp, err := h.service.LoginCashier(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем токен
	c.JSON(http.StatusOK, resp)
}

// logout обработчик для выхода из системы
func (h *Handler) logout(c *gin.Context) {
	// Получаем токен из заголовка
	token := c.GetHeader("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	// Выходим из системы
	if err := h.service.Logout(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}

// createCashier обработчик для создания кассира
func (h *Handler) createCashier(c *gin.Context) {
	var req models.CreateCashierRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем кассира
	resp, err := h.service.CreateCashier(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем созданного кассира
	c.JSON(http.StatusCreated, resp)
}

// getCashiers обработчик для получения списка кассиров
func (h *Handler) getCashiers(c *gin.Context) {
	// Получаем список кассиров
	resp, err := h.service.GetCashiers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем список кассиров
	c.JSON(http.StatusOK, resp)
}

// getCashierByID обработчик для получения кассира по ID
func (h *Handler) getCashierByID(c *gin.Context) {
	// Получаем ID кассира из URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID кассира"})
		return
	}

	// Получаем кассира по ID
	cashier, err := h.service.GetCashierByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем кассира
	c.JSON(http.StatusOK, gin.H{"cashier": cashier})
}

// updateCashier обработчик для обновления кассира
func (h *Handler) updateCashier(c *gin.Context) {
	// Получаем ID кассира из URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID кассира"})
		return
	}

	var req models.UpdateCashierRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем ID кассира
	req.ID = id

	// Обновляем кассира
	resp, err := h.service.UpdateCashier(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленного кассира
	c.JSON(http.StatusOK, resp)
}

// deleteCashier обработчик для удаления кассира
func (h *Handler) deleteCashier(c *gin.Context) {
	// Получаем ID кассира из URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID кассира"})
		return
	}

	// Удаляем кассира
	if err := h.service.DeleteCashier(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Кассир успешно удален"})
}

// authMiddleware middleware для проверки авторизации
func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		// Удаляем префикс "Bearer "
		token = strings.TrimPrefix(token, "Bearer ")

		// Проверяем токен
		claims, err := h.service.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.ID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)

		c.Next()
	}
}

// adminMiddleware middleware для проверки прав администратора
func (h *Handler) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Сначала проверяем авторизацию
		h.authMiddleware()(c)
		if c.IsAborted() {
			return
		}

		// Проверяем, является ли пользователь администратором
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Требуются права администратора"})
			c.Abort()
			return
		}

		c.Next()
	}
}
