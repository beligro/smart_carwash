package handlers

import (
	"net/http"
	"strings"

	"carwash_backend/internal/domain/auth/models"
	"carwash_backend/internal/domain/auth/service"
	"carwash_backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов авторизации
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(s service.Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes регистрирует маршруты для авторизации
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	authRoutes := router.Group("/auth")
	{
		// Маршруты для авторизации
		authRoutes.POST("/admin/login", h.loginAdmin)
		authRoutes.GET("/cashier/list", h.listCashiersForLogin)
		authRoutes.POST("/cashier/login", h.loginCashier)
		authRoutes.POST("/cleaner/login", h.loginCleaner)
		authRoutes.POST("/logout", h.authMiddleware(), h.logout)

		// Веб-авторизация по email
		webAuth := authRoutes.Group("/web")
		{
			webAuth.POST("/register/send-code", h.webRegisterSendCode)
			webAuth.POST("/register/verify", h.webRegisterVerify)
			webAuth.POST("/login", h.webLogin)
			webAuth.POST("/logout", h.webLogoutMiddleware(), h.webLogout)
			webAuth.POST("/change-password/send-code", h.webChangePasswordSendCode)
			webAuth.POST("/change-password/confirm", h.webChangePasswordConfirm)
		}

		// Маршруты для управления администраторами (только владелец, super_admin)
		adminRoutes := authRoutes.Group("/admins", h.superAdminMiddleware())
		{
			adminRoutes.POST("", h.createAdmin)
			adminRoutes.GET("", h.getAdmins)
			adminRoutes.PUT("", h.updateAdmin)
		}

		// Маршруты для управления кассирами (раздел cashiers)
		cashierRoutes := authRoutes.Group("/cashiers", h.requireSection("cashiers"))
		{
			cashierRoutes.POST("", h.createCashier)
			cashierRoutes.GET("", h.getCashiers)
			cashierRoutes.GET("/by-id", h.getCashierByID)
			cashierRoutes.PUT("", h.updateCashier)
			cashierRoutes.DELETE("", h.deleteCashier)
		}

		// Маршруты для управления уборщиками (раздел cleaners)
		cleanerRoutes := authRoutes.Group("/cleaners", h.requireSection("cleaners"))
		{
			cleanerRoutes.POST("", h.createCleaner)
			cleanerRoutes.GET("", h.getCleaners)
			cleanerRoutes.GET("/by-id", h.getCleanerByID)
			cleanerRoutes.PUT("", h.updateCleaner)
			cleanerRoutes.DELETE("", h.deleteCleaner)
		}

		// Маршруты для кассира (требуют авторизации кассира)
		cashierShiftRoutes := authRoutes.Group("/cashier", middleware.CashierMiddleware(h.service))
		{
			cashierShiftRoutes.POST("/shift/start", h.startShift)
			cashierShiftRoutes.POST("/shift/end", h.endShift)
			cashierShiftRoutes.GET("/shift/status", h.getShiftStatus)
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
	resp, err := h.service.LoginAdmin(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем токен
	c.JSON(http.StatusOK, resp)
}

// listCashiersForLogin возвращает имена активных кассиров для выпадающего списка на странице входа
func (h *Handler) listCashiersForLogin(c *gin.Context) {
	names, err := h.service.ListActiveCashierUsernames(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cashiers": names})
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
	resp, err := h.service.LoginCashier(c.Request.Context(), req.Username, req.Password)
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
	if err := h.service.Logout(c.Request.Context(), token); err != nil {
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
	resp, err := h.service.CreateCashier(c.Request.Context(), &req)
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
	resp, err := h.service.GetCashiers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем список кассиров
	c.JSON(http.StatusOK, resp)
}

// getCashierByID обработчик для получения кассира по ID
func (h *Handler) getCashierByID(c *gin.Context) {
	// Получаем ID кассира из query параметра
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID кассира обязателен"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID кассира"})
		return
	}

	// Получаем кассира по ID
	cashier, err := h.service.GetCashierByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"cashier_id": id,
		"username":   cashier.Username,
	})

	// Возвращаем кассира
	c.JSON(http.StatusOK, gin.H{"cashier": cashier})
}

// updateCashier обработчик для обновления кассира
func (h *Handler) updateCashier(c *gin.Context) {
	var req models.UpdateCashierRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем кассира
	resp, err := h.service.UpdateCashier(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"cashier_id": req.ID,
		"username":   resp.Username,
	})

	// Возвращаем обновленного кассира
	c.JSON(http.StatusOK, resp)
}

// deleteCashier обработчик для удаления кассира
func (h *Handler) deleteCashier(c *gin.Context) {
	var req models.DeleteCashierRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Удаляем кассира
	if err := h.service.DeleteCashier(c.Request.Context(), req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"cashier_id": req.ID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Кассир успешно удален"})
}

// authMiddleware middleware для проверки авторизации
// createAdmin создаёт персональную учётку администратора (только super_admin).
func (h *Handler) createAdmin(c *gin.Context) {
	var req models.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	admin, err := h.service.CreateAdmin(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, admin)
}

// getAdmins возвращает список админ-учёток (только super_admin).
func (h *Handler) getAdmins(c *gin.Context) {
	resp, err := h.service.GetAdmins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// updateAdmin обновляет учётку админа (только super_admin).
func (h *Handler) updateAdmin(c *gin.Context) {
	var req models.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	admin, err := h.service.UpdateAdmin(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, admin)
}

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
		claims, err := h.service.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.ID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)
		if claims.Role != "" {
			c.Set("role", claims.Role)
		}
		c.Set("allowed_sections", claims.AllowedSections)

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

// superAdminMiddleware пропускает только владельца (role=super_admin).
func (h *Handler) superAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.adminMiddleware()(c)
		if c.IsAborted() {
			return
		}
		role, _ := c.Get("role")
		if role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Требуются права владельца (super_admin)"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// requireSection ограничивает доступ по разделу. super_admin — всегда можно;
// легаси общий limited_admin (без allowed_sections) — тоже можно (совместимость
// до финального перехода); персональная учётка limited_admin — только если раздел
// есть в её allowed_sections.
func (h *Handler) requireSection(section string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.adminMiddleware()(c)
		if c.IsAborted() {
			return
		}
		role, _ := c.Get("role")
		if role == "super_admin" {
			c.Next()
			return
		}
		sectionsVal, _ := c.Get("allowed_sections")
		sections, _ := sectionsVal.([]string)
		if len(sections) == 0 {
			// Легаси общий limited_admin — полный доступ до перехода.
			c.Next()
			return
		}
		for _, s := range sections {
			if s == section {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к разделу"})
		c.Abort()
	}
}

// GetAdminMiddleware возвращает middleware для проверки прав администратора (публичный метод)
func (h *Handler) GetAdminMiddleware() gin.HandlerFunc {
	return h.adminMiddleware()
}

// cleanerMiddleware middleware для проверки авторизации уборщика
func (h *Handler) cleanerMiddleware() gin.HandlerFunc {
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

		// Проверяем токен уборщика
		claims, err := h.service.ValidateCleanerToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Сохраняем данные уборщика в контексте
		c.Set("cleaner_id", claims.ID)
		c.Set("username", claims.Username)
		c.Set("is_admin", false) // Уборщик не администратор

		c.Next()
	}
}

// GetCleanerMiddleware возвращает middleware для уборщиков
func (h *Handler) GetCleanerMiddleware() gin.HandlerFunc {
	return h.cleanerMiddleware()
}

// GetWebAuthMiddleware возвращает middleware для веб-клиента (JWT с role=web)
func (h *Handler) GetWebAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := h.service.ValidateWebToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("user_id", claims.ID)
		c.Set("role", "web")
		c.Next()
	}
}

// startShift обработчик для начала смены кассира
func (h *Handler) startShift(c *gin.Context) {
	// Получаем ID кассира из контекста (установлен middleware)
	cashierID, exists := c.Get("cashier_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	// Создаем запрос
	req := &models.StartShiftRequest{
		CashierID: cashierID.(uuid.UUID),
	}

	// Начинаем смену
	resp, err := h.service.StartShift(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"cashier_id": req.CashierID,
		"shift_id":   resp.ID,
	})

	// Возвращаем результат
	c.JSON(http.StatusOK, resp)
}

// endShift обработчик для завершения смены кассира
func (h *Handler) endShift(c *gin.Context) {
	// Получаем ID кассира из контекста (установлен middleware)
	cashierID, exists := c.Get("cashier_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	// Создаем запрос
	req := &models.EndShiftRequest{
		CashierID: cashierID.(uuid.UUID),
	}

	// Завершаем смену
	resp, err := h.service.EndShift(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"cashier_id": req.CashierID,
		"shift_id":   resp.ID,
	})

	// Возвращаем результат
	c.JSON(http.StatusOK, resp)
}

// getShiftStatus обработчик для получения статуса смены кассира
func (h *Handler) getShiftStatus(c *gin.Context) {
	// Получаем ID кассира из контекста (установлен middleware)
	cashierID, exists := c.Get("cashier_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	// Получаем статус смены
	resp, err := h.service.GetShiftStatus(c.Request.Context(), cashierID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"cashier_id": cashierID,
		"has_shift":  resp.HasActiveShift,
	})

	// Возвращаем результат
	c.JSON(http.StatusOK, resp)
}

// loginCleaner обработчик для авторизации уборщика
func (h *Handler) loginCleaner(c *gin.Context) {
	var req models.LoginRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Авторизуем уборщика
	resp, err := h.service.LoginCleaner(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем токен
	c.JSON(http.StatusOK, resp)
}

// createCleaner обработчик для создания уборщика
func (h *Handler) createCleaner(c *gin.Context) {
	var req models.CreateCleanerRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем уборщика
	resp, err := h.service.CreateCleaner(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusCreated, resp)
}

// getCleaners обработчик для получения списка уборщиков
func (h *Handler) getCleaners(c *gin.Context) {
	// Получаем список уборщиков
	resp, err := h.service.GetCleaners(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, resp)
}

// getCleanerByID обработчик для получения уборщика по ID
func (h *Handler) getCleanerByID(c *gin.Context) {
	// Получаем ID из query параметра
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID обязателен"})
		return
	}

	// Парсим UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
		return
	}

	// Получаем уборщика
	cleaner, err := h.service.GetCleanerByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, cleaner)
}

// updateCleaner обработчик для обновления уборщика
func (h *Handler) updateCleaner(c *gin.Context) {
	var req models.UpdateCleanerRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем уборщика
	resp, err := h.service.UpdateCleaner(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, resp)
}

// deleteCleaner обработчик для удаления уборщика
func (h *Handler) deleteCleaner(c *gin.Context) {
	var req models.DeleteCleanerRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Удаляем уборщика
	err := h.service.DeleteCleaner(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, gin.H{"message": "Уборщик удален успешно"})
}

// webRegisterSendCode отправляет код подтверждения на email при регистрации
func (h *Handler) webRegisterSendCode(c *gin.Context) {
	var req models.WebRegisterSendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.RegisterSendCode(c.Request.Context(), req.Email, req.Password, req.PasswordConfirm); err != nil {
		if err == service.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// webRegisterVerify проверяет код и создаёт пользователя, возвращает JWT
func (h *Handler) webRegisterVerify(c *gin.Context) {
	var req models.WebRegisterVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.service.VerifyEmailAndRegister(c.Request.Context(), req.Email, req.Code, req.Password, req.PasswordConfirm)
	if err != nil {
		if err == service.ErrInvalidVerificationCode {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// webLogin вход по email и паролю
func (h *Handler) webLogin(c *gin.Context) {
	var req models.WebLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.service.LoginWeb(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrWebAuthNotConfigured {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// webLogoutMiddleware проверяет веб-JWT и устанавливает user_id в контексте
func (h *Handler) webLogoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := h.service.ValidateWebToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("user_id", claims.ID)
		c.Set("role", "web")
		c.Set("_web_token", token)
		c.Next()
	}
}

// webLogout инвалидирует веб-токен
func (h *Handler) webLogout(c *gin.Context) {
	token, _ := c.Get("_web_token")
	if tokenStr, ok := token.(string); ok {
		h.service.InvalidateWebToken(tokenStr)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}

// webChangePasswordSendCode отправляет код на email для смены пароля
func (h *Handler) webChangePasswordSendCode(c *gin.Context) {
	var req models.WebChangePasswordSendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.ChangePasswordSendCode(c.Request.Context(), req.Email); err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь с таким email не найден"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// webChangePasswordConfirm смена пароля после ввода кода
func (h *Handler) webChangePasswordConfirm(c *gin.Context) {
	var req models.WebChangePasswordConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.ChangePasswordConfirm(c.Request.Context(), req.Email, req.Code, req.NewPassword, req.NewPasswordConfirm); err != nil {
		if err == service.ErrInvalidVerificationCode {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
