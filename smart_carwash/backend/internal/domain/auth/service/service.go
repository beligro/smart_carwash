package service

import (
	"carwash_backend/internal/logger"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/auth/models"
	"carwash_backend/internal/domain/auth/repository"
	userModels "carwash_backend/internal/domain/user/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// ErrInvalidCredentials возвращается при неверных учетных данных
	ErrInvalidCredentials = errors.New("неверное имя пользователя или пароль")

	// ErrCashierInactive возвращается, когда кассир неактивен
	ErrCashierInactive = errors.New("кассир неактивен")

	// ErrAnotherCashierActive возвращается, когда другой кассир уже активен
	ErrAnotherCashierActive = errors.New("другой кассир уже активен в системе")

	// ErrCleanerInactive возвращается, когда уборщик неактивен
	ErrCleanerInactive = errors.New("уборщик неактивен")

	// ErrAnotherCleanerActive возвращается, когда другой уборщик уже активен
	ErrAnotherCleanerActive = errors.New("другой уборщик уже активен в системе")

	// ErrAdminAlreadyExists возвращается при попытке создать второго администратора
	ErrAdminAlreadyExists = errors.New("администратор уже существует")

	// ErrWebAuthNotConfigured веб-авторизация не настроена
	ErrWebAuthNotConfigured = errors.New("веб-авторизация не настроена")
	// ErrUserAlreadyExists пользователь с таким email уже зарегистрирован
	ErrUserAlreadyExists = errors.New("пользователь с таким email уже зарегистрирован")
	// ErrInvalidVerificationCode неверный или истёкший код подтверждения
	ErrInvalidVerificationCode = errors.New("неверный или истёкший код подтверждения")
)

var (
	emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// Service интерфейс для бизнес-логики авторизации
type Service interface {
	// Методы для авторизации
	LoginAdmin(ctx context.Context, username, password string) (*models.LoginResponse, error)
	LoginCashier(ctx context.Context, username, password string) (*models.LoginResponse, error)
	ListActiveCashierUsernames(ctx context.Context) ([]string, error)
	LoginCleaner(ctx context.Context, username, password string) (*models.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*models.TokenClaims, error)
	ValidateCleanerToken(ctx context.Context, token string) (*models.TokenClaims, error)
	Logout(ctx context.Context, token string) error

	// Методы для управления кассирами
	CreateCashier(ctx context.Context, req *models.CreateCashierRequest) (*models.CreateCashierResponse, error)
	UpdateCashier(ctx context.Context, req *models.UpdateCashierRequest) (*models.UpdateCashierResponse, error)
	DeleteCashier(ctx context.Context, id uuid.UUID) error
	GetCashiers(ctx context.Context) (*models.GetCashiersResponse, error)
	GetCashierByID(ctx context.Context, id uuid.UUID) (*models.Cashier, error)

	// Методы для управления администраторами (персональные учётки)
	CreateAdmin(ctx context.Context, req *models.CreateAdminRequest) (*models.Admin, error)
	UpdateAdmin(ctx context.Context, req *models.UpdateAdminRequest) (*models.Admin, error)
	GetAdmins(ctx context.Context) (*models.GetAdminsResponse, error)

	// Методы для управления сменами кассиров
	StartShift(ctx context.Context, req *models.StartShiftRequest) (*models.StartShiftResponse, error)
	EndShift(ctx context.Context, req *models.EndShiftRequest) (*models.EndShiftResponse, error)
	GetShiftStatus(ctx context.Context, cashierID uuid.UUID) (*models.ShiftStatusResponse, error)

	// Методы для управления уборщиками
	CreateCleaner(ctx context.Context, req *models.CreateCleanerRequest) (*models.CreateCleanerResponse, error)
	UpdateCleaner(ctx context.Context, req *models.UpdateCleanerRequest) (*models.UpdateCleanerResponse, error)
	DeleteCleaner(ctx context.Context, id uuid.UUID) error
	GetCleaners(ctx context.Context) (*models.GetCleanersResponse, error)
	GetCleanerByID(ctx context.Context, id uuid.UUID) (*models.Cleaner, error)

	// Методы для двухфакторной аутентификации (заглушки для будущей реализации)
	EnableTwoFactorAuth(ctx context.Context, userID uuid.UUID) (string, error)
	DisableTwoFactorAuth(ctx context.Context, userID uuid.UUID) error
	VerifyTwoFactorCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)

	// Веб-авторизация по email (JWT для веб-клиента)
	GenerateWebToken(userID uuid.UUID) (string, time.Time, error)
	ValidateWebToken(tokenString string) (*models.TokenClaims, error)
	InvalidateWebToken(tokenString string)

	// Регистрация и вход по email
	RegisterSendCode(ctx context.Context, email, password, passwordConfirm string) error
	VerifyEmailAndRegister(ctx context.Context, email, code, password, passwordConfirm string) (*models.WebAuthResponse, error)
	LoginWeb(ctx context.Context, email, password string) (*models.WebAuthResponse, error)

	// Смена пароля по коду на email
	ChangePasswordSendCode(ctx context.Context, email string) error
	ChangePasswordConfirm(ctx context.Context, email, code, newPassword, newPasswordConfirm string) error
}

// WebUserAuth интерфейс для создания/обновления веб-пользователей (user service)
type WebUserAuth interface {
	GetUserByEmail(ctx context.Context, email string) (*userModels.User, error)
	CreateWebUser(ctx context.Context, email, passwordHash string) (*userModels.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

// WebEmailCodeSender отправка и проверка кодов на email
type WebEmailCodeSender interface {
	SendRegisterCode(ctx context.Context, email string) error
	VerifyRegisterCode(email, code string) bool
	SendPasswordChangeCode(ctx context.Context, email string) error
	VerifyPasswordChangeCode(email, code string) bool
}

// invalidTokenCacheEntry запись в кэше невалидных токенов
type invalidTokenCacheEntry struct {
	expiresAt time.Time
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo             repository.Repository
	config           *config.Config
	webUserAuth      WebUserAuth      // опционально, для веб-авторизации
	webEmailCodeSender WebEmailCodeSender // опционально
	invalidTokenCache sync.Map
	cacheTTL          time.Duration
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, config *config.Config, webUserAuth WebUserAuth, webEmailCodeSender WebEmailCodeSender) *ServiceImpl {
	service := &ServiceImpl{
		repo:                repo,
		config:              config,
		webUserAuth:         webUserAuth,
		webEmailCodeSender:  webEmailCodeSender,
		cacheTTL:            5 * time.Minute,
	}

	// Запускаем фоновую очистку кэша
	go service.cleanupInvalidTokenCache()

	return service
}

// LoginAdmin авторизует администратора: super_admin и общий limited_admin из конфига,
// либо персональные учётки из таблицы admins (роль limited_admin + allowed_sections).
func (s *ServiceImpl) LoginAdmin(ctx context.Context, username, password string) (*models.LoginResponse, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	role := ""
	var allowedSections []string
	adminID := uuid.New() // для конфиг-учёток генерируем случайный ID

	switch {
	case username == s.config.AdminUsername && password == s.config.AdminPassword:
		role = "super_admin"
	case s.config.LimitedAdminUsername != "" && s.config.LimitedAdminPassword != "" &&
		username == s.config.LimitedAdminUsername && password == s.config.LimitedAdminPassword:
		// Легаси общий limited_admin (оставляем рабочим до финального перехода): без allowed_sections.
		role = "limited_admin"
	default:
		// Персональная учётка админа из БД
		admin, err := s.repo.GetAdminByUsername(ctx, username)
		if err != nil {
			if errors.Is(err, repository.ErrAdminNotFound) {
				return nil, ErrInvalidCredentials
			}
			return nil, err
		}
		if !admin.IsActive {
			return nil, ErrInvalidCredentials
		}
		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
			return nil, ErrInvalidCredentials
		}
		role = "limited_admin"
		allowedSections = []string(admin.AllowedSections)
		adminID = admin.ID
		now := time.Now()
		admin.LastLogin = &now
		_ = s.repo.UpdateAdmin(ctx, admin)
	}

	claims := models.TokenClaims{
		ID:              adminID,
		Username:        username,
		IsAdmin:         true,
		Role:            role,
		AllowedSections: allowedSections,
	}

	token, expiresAt, err := s.generateToken(claims)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:           token,
		ExpiresAt:       expiresAt,
		IsAdmin:         true,
		Role:            role,
		AllowedSections: allowedSections,
	}, nil
}

// CreateAdmin создаёт персональную учётку администратора.
func (s *ServiceImpl) CreateAdmin(ctx context.Context, req *models.CreateAdminRequest) (*models.Admin, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	admin := &models.Admin{
		Username:        strings.TrimSpace(req.Username),
		PasswordHash:    string(hashed),
		DisplayName:     req.DisplayName,
		AllowedSections: models.Sections(req.AllowedSections),
		IsActive:        true,
	}
	if err := s.repo.CreateAdmin(ctx, admin); err != nil {
		if errors.Is(err, repository.ErrAdminAlreadyExists) {
			return nil, fmt.Errorf("администратор с именем %s уже существует", req.Username)
		}
		return nil, err
	}
	return admin, nil
}

// UpdateAdmin обновляет учётку админа (пароль/имя/разделы/активность).
func (s *ServiceImpl) UpdateAdmin(ctx context.Context, req *models.UpdateAdminRequest) (*models.Admin, error) {
	admin, err := s.repo.GetAdminByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		admin.PasswordHash = string(hashed)
	}
	if req.DisplayName != "" {
		admin.DisplayName = req.DisplayName
	}
	if req.AllowedSections != nil {
		admin.AllowedSections = models.Sections(req.AllowedSections)
	}
	if req.IsActive != nil {
		admin.IsActive = *req.IsActive
	}
	if err := s.repo.UpdateAdmin(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

// GetAdmins возвращает список персональных админ-учёток.
func (s *ServiceImpl) GetAdmins(ctx context.Context) (*models.GetAdminsResponse, error) {
	admins, err := s.repo.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}
	return &models.GetAdminsResponse{Admins: admins}, nil
}

// ListActiveCashierUsernames возвращает имена активных кассиров (для выпадающего списка на логине).
func (s *ServiceImpl) ListActiveCashierUsernames(ctx context.Context) ([]string, error) {
	cashiers, err := s.repo.ListCashiers(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cashiers))
	for _, c := range cashiers {
		if c.IsActive {
			names = append(names, c.Username)
		}
	}
	return names, nil
}

// LoginCashier авторизует кассира
func (s *ServiceImpl) LoginCashier(ctx context.Context, username, password string) (*models.LoginResponse, error) {
	// Убираем случайные пробелы/переносы (частая причина ложных ошибок при вводе на терминале)
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	logger.Printf("Попытка входа кассира: username=%s", username)

	// Получаем кассира по имени пользователя
	cashier, err := s.repo.GetCashierByUsername(ctx, username)
	if err != nil {
		logger.Printf("Ошибка получения кассира: %v", err)
		if errors.Is(err, repository.ErrCashierNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	logger.Printf("Кассир найден: ID=%s, Username=%s, IsActive=%t", cashier.ID, cashier.Username, cashier.IsActive)

	// Проверяем, активен ли кассир
	if !cashier.IsActive {
		logger.Printf("Кассир неактивен: ID=%s", cashier.ID)
		return nil, ErrCashierInactive
	}

	// Проверяем пароль
	logger.Printf("Проверка пароля для кассира: ID=%s", cashier.ID)
	if err := bcrypt.CompareHashAndPassword([]byte(cashier.PasswordHash), []byte(password)); err != nil {
		logger.Printf("Неверный пароль для кассира: ID=%s, error=%v", cashier.ID, err)
		return nil, ErrInvalidCredentials
	}

	logger.Printf("Пароль верный для кассира: ID=%s", cashier.ID)

	// Создаем JWT токен для кассира
	claims := models.TokenClaims{
		ID:       cashier.ID,
		Username: cashier.Username,
		IsAdmin:  false,
	}

	// Генерируем токен
	token, expiresAt, err := s.generateToken(claims)
	if err != nil {
		return nil, err
	}

	// Создаем сессию кассира
	session := &models.CashierSession{
		CashierID: cashier.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	// Сохраняем сессию в базе данных
	if err := s.repo.CreateCashierSession(ctx, session); err != nil {
		if errors.Is(err, repository.ErrActiveCashierSessionExists) {
			return nil, ErrAnotherCashierActive
		}
		return nil, err
	}

	// Обновляем время последнего входа кассира
	now := time.Now()
	cashier.LastLogin = &now
	if err := s.repo.UpdateCashier(ctx, cashier); err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		IsAdmin:   false,
	}, nil
}

// addToInvalidTokenCache добавляет токен в кэш невалидных токенов
func (s *ServiceImpl) addToInvalidTokenCache(tokenString string) {
	entry := invalidTokenCacheEntry{
		expiresAt: time.Now().Add(s.cacheTTL),
	}
	s.invalidTokenCache.Store(tokenString, entry)
}

// removeFromInvalidTokenCache удаляет токен из кэша невалидных токенов
// Используется при логауте или успешном логине
func (s *ServiceImpl) removeFromInvalidTokenCache(tokenString string) {
	s.invalidTokenCache.Delete(tokenString)
}

// isInInvalidTokenCache проверяет, есть ли токен в кэше невалидных токенов
func (s *ServiceImpl) isInInvalidTokenCache(tokenString string) bool {
	value, ok := s.invalidTokenCache.Load(tokenString)
	if !ok {
		return false
	}

	entry, ok := value.(invalidTokenCacheEntry)
	if !ok {
		// Невалидная запись, удаляем
		s.invalidTokenCache.Delete(tokenString)
		return false
	}

	// Проверяем, не истек ли TTL
	if time.Now().After(entry.expiresAt) {
		// TTL истек, удаляем запись
		s.invalidTokenCache.Delete(tokenString)
		return false
	}

	return true
}

// cleanupInvalidTokenCache периодически очищает истекшие записи из кэша
// Вызывается в фоне при старте сервиса
func (s *ServiceImpl) cleanupInvalidTokenCache() {
	for {
		time.Sleep(1 * time.Minute) // Очистка каждую минуту

		now := time.Now()
		s.invalidTokenCache.Range(func(key, value interface{}) bool {
			entry, ok := value.(invalidTokenCacheEntry)
			if !ok || now.After(entry.expiresAt) {
				// Запись истекла, удаляем
				s.invalidTokenCache.Delete(key)
			}
			return true
		})
	}
}

// ValidateToken проверяет JWT токен и возвращает данные пользователя
func (s *ServiceImpl) ValidateToken(ctx context.Context, tokenString string) (*models.TokenClaims, error) {
	// Сначала проверяем кэш невалидных токенов
	if s.isInInvalidTokenCache(tokenString) {
		return nil, errors.New("токен недействителен")
	}

	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		// Токен невалидный, добавляем в кэш
		s.addToInvalidTokenCache(tokenString)
		return nil, err
	}

	// Проверяем валидность токена
	if !token.Valid {
		// Токен невалидный, добавляем в кэш
		s.addToInvalidTokenCache(tokenString)
		return nil, errors.New("недействительный токен")
	}

	// Получаем данные из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// Токен невалидный, добавляем в кэш
		s.addToInvalidTokenCache(tokenString)
		return nil, errors.New("недействительные данные токена")
	}

	// Проверяем, не истек ли токен
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			// Токен истек, добавляем в кэш
			s.addToInvalidTokenCache(tokenString)
			return nil, errors.New("токен истек")
		}
	}

	// Если это токен кассира, проверяем, существует ли сессия
	isAdmin, _ := claims["is_admin"].(bool)
	if !isAdmin {
		// Проверяем контекст перед DB запросом
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Создаем отдельный контекст с коротким таймаутом для DB запроса (2 секунды)
		// Это предотвратит зависание на долгих запросах к БД
		dbCtx, dbCancel := context.WithTimeout(ctx, 2*time.Second)
		defer dbCancel()

		// Получаем сессию по токену
		session, err := s.repo.GetCashierSessionByToken(dbCtx, tokenString)
		
		// Проверяем, был ли отменен контекст или превышен таймаут
		if dbCtx.Err() != nil {
			return nil, dbCtx.Err()
		}

		if err != nil {
			// Ошибка при запросе к БД, не добавляем в кэш (может быть временная ошибка)
			return nil, err
		}
		if session == nil {
			// Сессия не найдена, добавляем в кэш невалидных токенов
			s.addToInvalidTokenCache(tokenString)
			return nil, errors.New("сессия не найдена или истекла")
		}
	}

	// Преобразуем данные в структуру TokenClaims
	idStr, _ := claims["id"].(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	username, _ := claims["username"].(string)

	role, _ := claims["role"].(string)

	var allowedSections []string
	if raw, ok := claims["allowed_sections"].([]interface{}); ok {
		for _, v := range raw {
			if str, ok := v.(string); ok {
				allowedSections = append(allowedSections, str)
			}
		}
	}

	return &models.TokenClaims{
		ID:              id,
		Username:        username,
		IsAdmin:         isAdmin,
		Role:            role,
		AllowedSections: allowedSections,
	}, nil
}

// ValidateCleanerToken проверяет токен уборщика и возвращает данные из него
func (s *ServiceImpl) ValidateCleanerToken(ctx context.Context, tokenString string) (*models.TokenClaims, error) {
	// Сначала проверяем кэш невалидных токенов
	if s.isInInvalidTokenCache(tokenString) {
		return nil, errors.New("токен недействителен")
	}

	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		// Токен невалидный, добавляем в кэш
		s.addToInvalidTokenCache(tokenString)
		return nil, err
	}

	// Проверяем валидность токена
	if !token.Valid {
		// Токен невалидный, добавляем в кэш
		s.addToInvalidTokenCache(tokenString)
		return nil, errors.New("недействительный токен")
	}

	// Получаем данные из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// Токен невалидный, добавляем в кэш
		s.addToInvalidTokenCache(tokenString)
		return nil, errors.New("недействительные данные токена")
	}

	// Проверяем, не истек ли токен
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			// Токен истек, добавляем в кэш
			s.addToInvalidTokenCache(tokenString)
			return nil, errors.New("токен истек")
		}
	}

	// Проверяем, что это токен уборщика (не администратора)
	isAdmin, _ := claims["is_admin"].(bool)
	if isAdmin {
		// Токен не для уборщика, добавляем в кэш
		s.addToInvalidTokenCache(tokenString)
		return nil, errors.New("этот токен не для уборщика")
	}

	// Проверяем контекст перед DB запросом
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Создаем отдельный контекст с коротким таймаутом для DB запроса (2 секунды)
	// Это предотвратит зависание на долгих запросах к БД
	dbCtx, dbCancel := context.WithTimeout(ctx, 2*time.Second)
	defer dbCancel()

	// Получаем сессию уборщика по токену
	session, err := s.repo.GetCleanerSessionByToken(dbCtx, tokenString)
	
	// Проверяем, был ли отменен контекст или превышен таймаут
	if dbCtx.Err() != nil {
		return nil, dbCtx.Err()
	}

	if err != nil {
		// Ошибка при запросе к БД, не добавляем в кэш (может быть временная ошибка)
		return nil, err
	}
	if session == nil {
		// Сессия не найдена, добавляем в кэш невалидных токенов
		s.addToInvalidTokenCache(tokenString)
		return nil, errors.New("сессия уборщика не найдена или истекла")
	}

	// Преобразуем данные в структуру TokenClaims
	idStr, _ := claims["id"].(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	username, _ := claims["username"].(string)

	return &models.TokenClaims{
		ID:       id,
		Username: username,
		IsAdmin:  false, // Уборщик никогда не администратор
	}, nil
}

// Logout завершает сессию пользователя
func (s *ServiceImpl) Logout(ctx context.Context, token string) error {
	// Проверяем токен
	claims, err := s.ValidateToken(ctx, token)
	if err != nil {
		return err
	}

	// Удаляем токен из кэша невалидных токенов (если был там)
	s.removeFromInvalidTokenCache(token)

	// Если это администратор, просто возвращаем успех
	if claims.IsAdmin {
		return nil
	}

	// Если это кассир, удаляем сессию
	session, err := s.repo.GetCashierSessionByToken(ctx, token)
	if err != nil {
		return err
	}
	if session == nil {
		return nil // Сессия уже удалена или истекла
	}

	// Логаут кассира закрывает его активную смену (передача смены: «Выйти» = смена
	// закрыта, следующий кассир заходит под собой). Закрываем только смену этого кассира.
	if shift, sErr := s.repo.GetActiveCashierShift(ctx); sErr == nil && shift != nil && shift.CashierID == session.CashierID {
		now := time.Now()
		shift.EndedAt = &now
		shift.IsActive = false
		if uErr := s.repo.UpdateCashierShift(ctx, shift); uErr != nil {
			logger.Printf("Logout: не удалось завершить смену кассира %s при выходе: %v", session.CashierID, uErr)
		}
	}

	return s.repo.DeleteCashierSession(ctx, session.ID)
}

// CreateCashier создает нового кассира
func (s *ServiceImpl) CreateCashier(ctx context.Context, req *models.CreateCashierRequest) (*models.CreateCashierResponse, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем нового кассира
	cashier := &models.Cashier{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}

	// Сохраняем кассира в базе данных
	if err := s.repo.CreateCashier(ctx, cashier); err != nil {
		if errors.Is(err, repository.ErrCashierAlreadyExists) {
			return nil, fmt.Errorf("кассир с именем %s уже существует", req.Username)
		}
		return nil, err
	}

	return &models.CreateCashierResponse{
		ID:        cashier.ID,
		Username:  cashier.Username,
		CreatedAt: cashier.CreatedAt,
	}, nil
}

// UpdateCashier обновляет кассира
func (s *ServiceImpl) UpdateCashier(ctx context.Context, req *models.UpdateCashierRequest) (*models.UpdateCashierResponse, error) {
	// Получаем кассира по ID
	cashier, err := s.repo.GetCashierByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrCashierNotFound) {
			return nil, fmt.Errorf("кассир с ID %s не найден", req.ID)
		}
		return nil, err
	}

	// Обновляем данные кассира
	if req.Username != "" {
		cashier.Username = req.Username
	}

	if req.Password != "" {
		// Хешируем новый пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		cashier.PasswordHash = string(hashedPassword)
	}

	cashier.IsActive = req.IsActive

	// Сохраняем обновленного кассира
	if err := s.repo.UpdateCashier(ctx, cashier); err != nil {
		return nil, err
	}

	return &models.UpdateCashierResponse{
		ID:        cashier.ID,
		Username:  cashier.Username,
		IsActive:  cashier.IsActive,
		UpdatedAt: cashier.UpdatedAt,
	}, nil
}

// DeleteCashier удаляет кассира
func (s *ServiceImpl) DeleteCashier(ctx context.Context, id uuid.UUID) error {
	// Проверяем, существует ли кассир
	_, err := s.repo.GetCashierByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCashierNotFound) {
			return fmt.Errorf("кассир с ID %s не найден", id)
		}
		return err
	}

	// Удаляем все сессии кассира
	if err := s.repo.DeleteCashierSession(ctx, id); err != nil {
		return err
	}

	// Удаляем кассира
	return s.repo.DeleteCashier(ctx, id)
}

// GetCashiers возвращает список всех кассиров
func (s *ServiceImpl) GetCashiers(ctx context.Context) (*models.GetCashiersResponse, error) {
	cashiers, err := s.repo.ListCashiers(ctx)
	if err != nil {
		return nil, err
	}

	return &models.GetCashiersResponse{
		Cashiers: cashiers,
	}, nil
}

// GetCashierByID возвращает кассира по ID
func (s *ServiceImpl) GetCashierByID(ctx context.Context, id uuid.UUID) (*models.Cashier, error) {
	return s.repo.GetCashierByID(ctx, id)
}

// StartShift начинает смену для кассира
func (s *ServiceImpl) StartShift(ctx context.Context, req *models.StartShiftRequest) (*models.StartShiftResponse, error) {
	// Проверяем, существует ли кассир
	cashier, err := s.repo.GetCashierByID(ctx, req.CashierID)
	if err != nil {
		if errors.Is(err, repository.ErrCashierNotFound) {
			return nil, fmt.Errorf("кассир с ID %s не найден", req.CashierID)
		}
		return nil, err
	}

	// Проверяем, активен ли кассир
	if !cashier.IsActive {
		return nil, ErrCashierInactive
	}

	// Создаем новую смену
	// Смена истекает в 09:00 по Новосибирску (UTC+7) = 02:00 UTC каждый день.
	// Находим ближайшее 02:00 UTC, которое строго позже now.
	now := time.Now().UTC()
	todayExpire := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, time.UTC)
	if !now.Before(todayExpire) {
		todayExpire = todayExpire.Add(24 * time.Hour)
	}
	expiresAt := todayExpire

	shift := &models.CashierShift{
		CashierID: req.CashierID,
		StartedAt: now,
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	if err := s.repo.CreateCashierShift(ctx, shift); err != nil {
		if errors.Is(err, repository.ErrActiveShiftExists) {
			return nil, fmt.Errorf("уже есть активная смена")
		}
		return nil, err
	}

	return &models.StartShiftResponse{
		ID:        shift.ID,
		StartedAt: shift.StartedAt,
		ExpiresAt: shift.ExpiresAt,
		IsActive:  shift.IsActive,
	}, nil
}

// EndShift завершает смену для кассира
func (s *ServiceImpl) EndShift(ctx context.Context, req *models.EndShiftRequest) (*models.EndShiftResponse, error) {
	// Проверяем, существует ли кассир
	cashier, err := s.repo.GetCashierByID(ctx, req.CashierID)
	if err != nil {
		if errors.Is(err, repository.ErrCashierNotFound) {
			return nil, fmt.Errorf("кассир с ID %s не найден", req.CashierID)
		}
		return nil, err
	}

	// Проверяем, активен ли кассир
	if !cashier.IsActive {
		return nil, ErrCashierInactive
	}

	// Получаем активную смену
	shift, err := s.repo.GetActiveCashierShift(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveShift) {
			return nil, fmt.Errorf("нет активной смены")
		}
		return nil, err
	}

	// Проверяем, что смена принадлежит этому кассиру
	if shift.CashierID != req.CashierID {
		return nil, fmt.Errorf("активная смена принадлежит другому кассиру")
	}

	// Завершаем смену
	now := time.Now()
	shift.EndedAt = &now
	shift.IsActive = false

	if err := s.repo.UpdateCashierShift(ctx, shift); err != nil {
		return nil, err
	}

	return &models.EndShiftResponse{
		ID:        shift.ID,
		StartedAt: shift.StartedAt,
		EndedAt:   *shift.EndedAt,
		IsActive:  shift.IsActive,
	}, nil
}

// GetShiftStatus возвращает статус смены для кассира
func (s *ServiceImpl) GetShiftStatus(ctx context.Context, cashierID uuid.UUID) (*models.ShiftStatusResponse, error) {
	// Проверяем, существует ли кассир
	cashier, err := s.repo.GetCashierByID(ctx, cashierID)
	if err != nil {
		if errors.Is(err, repository.ErrCashierNotFound) {
			return nil, fmt.Errorf("кассир с ID %s не найден", cashierID)
		}
		return nil, err
	}

	// Проверяем, активен ли кассир
	if !cashier.IsActive {
		return nil, ErrCashierInactive
	}

	// Получаем активную смену
	shift, err := s.repo.GetActiveCashierShift(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveShift) {
			return &models.ShiftStatusResponse{
				HasActiveShift: false,
				Shift:          nil,
			}, nil
		}
		return nil, err
	}

	// Проверяем, что смена принадлежит этому кассиру
	if shift.CashierID != cashierID {
		return &models.ShiftStatusResponse{
			HasActiveShift: false,
			Shift:          nil,
		}, nil
	}

	return &models.ShiftStatusResponse{
		HasActiveShift: true,
		Shift:          shift,
	}, nil
}

// EnableTwoFactorAuth включает двухфакторную аутентификацию для пользователя
// (заглушка для будущей реализации)
func (s *ServiceImpl) EnableTwoFactorAuth(ctx context.Context, userID uuid.UUID) (string, error) {
	// Здесь будет логика генерации секрета и QR-кода для двухфакторной аутентификации
	// Пока просто заглушка
	settings := &models.TwoFactorAuthSettings{
		UserID:    userID,
		IsEnabled: true,
		Secret:    "dummy_secret", // В реальной реализации здесь будет настоящий секрет
	}

	if err := s.repo.SaveTwoFactorAuthSettings(ctx, settings); err != nil {
		return "", err
	}

	return "dummy_qr_code", nil // В реальной реализации здесь будет URL для QR-кода
}

// DisableTwoFactorAuth отключает двухфакторную аутентификацию для пользователя
// (заглушка для будущей реализации)
func (s *ServiceImpl) DisableTwoFactorAuth(ctx context.Context, userID uuid.UUID) error {
	settings := &models.TwoFactorAuthSettings{
		UserID:    userID,
		IsEnabled: false,
		Secret:    "",
	}

	return s.repo.SaveTwoFactorAuthSettings(ctx, settings)
}

// VerifyTwoFactorCode проверяет код двухфакторной аутентификации
// (заглушка для будущей реализации)
func (s *ServiceImpl) VerifyTwoFactorCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	// Здесь будет логика проверки кода двухфакторной аутентификации
	// Пока просто заглушка
	settings, err := s.repo.GetTwoFactorAuthSettings(ctx, userID)
	if err != nil {
		return false, err
	}

	if !settings.IsEnabled {
		return true, nil // Если двухфакторная аутентификация отключена, считаем код верным
	}

	// В реальной реализации здесь будет проверка кода
	return code == "123456", nil
}

// generateToken генерирует JWT токен
func (s *ServiceImpl) generateToken(claims models.TokenClaims) (string, time.Time, error) {
	// Устанавливаем время истечения токена (24 часа)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Создаем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":               claims.ID.String(),
		"username":         claims.Username,
		"is_admin":         claims.IsAdmin,
		"role":             claims.Role,
		"allowed_sections": claims.AllowedSections,
		"exp":              expiresAt.Unix(),
	})

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// LoginCleaner авторизует уборщика
func (s *ServiceImpl) LoginCleaner(ctx context.Context, username, password string) (*models.LoginResponse, error) {
	logger.Printf("Попытка входа уборщика: username=%s", username)

	// Получаем уборщика по имени пользователя
	cleaner, err := s.repo.GetCleanerByUsername(ctx, username)
	if err != nil {
		logger.Printf("Ошибка получения уборщика: %v", err)
		if errors.Is(err, repository.ErrCleanerNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	logger.Printf("Уборщик найден: ID=%s, Username=%s, IsActive=%t", cleaner.ID, cleaner.Username, cleaner.IsActive)

	// Проверяем, активен ли уборщик
	if !cleaner.IsActive {
		logger.Printf("Уборщик неактивен: ID=%s", cleaner.ID)
		return nil, ErrCleanerInactive
	}

	// Проверяем пароль
	logger.Printf("Проверка пароля для уборщика: ID=%s", cleaner.ID)
	if err := bcrypt.CompareHashAndPassword([]byte(cleaner.PasswordHash), []byte(password)); err != nil {
		logger.Printf("Неверный пароль для уборщика: ID=%s, error=%v", cleaner.ID, err)
		return nil, ErrInvalidCredentials
	}

	logger.Printf("Пароль верный для уборщика: ID=%s", cleaner.ID)

	// Создаем JWT токен для уборщика
	claims := models.TokenClaims{
		ID:       cleaner.ID,
		Username: cleaner.Username,
		IsAdmin:  false,
	}

	// Генерируем токен
	token, expiresAt, err := s.generateToken(claims)
	if err != nil {
		return nil, err
	}

	// Создаем сессию уборщика
	session := &models.CleanerSession{
		CleanerID: cleaner.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	// Сохраняем сессию в базе данных
	if err := s.repo.CreateCleanerSession(ctx, session); err != nil {
		return nil, err
	}

	// Обновляем время последнего входа уборщика
	now := time.Now()
	cleaner.LastLogin = &now
	if err := s.repo.UpdateCleaner(ctx, cleaner); err != nil {
		logger.Printf("Ошибка обновления времени последнего входа уборщика: %v", err)
		// Не возвращаем ошибку, так как авторизация прошла успешно
	}

	logger.Printf("Успешная авторизация уборщика: ID=%s, Username=%s", cleaner.ID, cleaner.Username)

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		IsAdmin:   false,
	}, nil
}

// CreateCleaner создает нового уборщика
func (s *ServiceImpl) CreateCleaner(ctx context.Context, req *models.CreateCleanerRequest) (*models.CreateCleanerResponse, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем уборщика
	cleaner := &models.Cleaner{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}

	// Сохраняем уборщика в базе данных
	err = s.repo.CreateCleaner(ctx, cleaner)
	if err != nil {
		return nil, err
	}

	return &models.CreateCleanerResponse{
		ID:        cleaner.ID,
		Username:  cleaner.Username,
		CreatedAt: cleaner.CreatedAt,
	}, nil
}

// UpdateCleaner обновляет уборщика
func (s *ServiceImpl) UpdateCleaner(ctx context.Context, req *models.UpdateCleanerRequest) (*models.UpdateCleanerResponse, error) {
	// Получаем существующего уборщика
	cleaner, err := s.repo.GetCleanerByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	cleaner.Username = req.Username
	cleaner.IsActive = req.IsActive

	// Если передан новый пароль, хешируем его
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		cleaner.PasswordHash = string(hashedPassword)
	}

	// Сохраняем изменения
	err = s.repo.UpdateCleaner(ctx, cleaner)
	if err != nil {
		return nil, err
	}

	return &models.UpdateCleanerResponse{
		ID:        cleaner.ID,
		Username:  cleaner.Username,
		IsActive:  cleaner.IsActive,
		UpdatedAt: cleaner.UpdatedAt,
	}, nil
}

// DeleteCleaner удаляет уборщика
func (s *ServiceImpl) DeleteCleaner(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteCleaner(ctx, id)
}

// GetCleaners получает список всех уборщиков
func (s *ServiceImpl) GetCleaners(ctx context.Context) (*models.GetCleanersResponse, error) {
	cleaners, err := s.repo.ListCleaners(ctx)
	if err != nil {
		return nil, err
	}

	return &models.GetCleanersResponse{
		Cleaners: cleaners,
	}, nil
}

// GetCleanerByID получает уборщика по ID
func (s *ServiceImpl) GetCleanerByID(ctx context.Context, id uuid.UUID) (*models.Cleaner, error) {
	return s.repo.GetCleanerByID(ctx, id)
}

// GenerateWebToken генерирует JWT для веб-клиента (aud=web, id=user_id). Срок жизни 7 дней, чтобы пользователь не вылетал из аккаунта.
func (s *ServiceImpl) GenerateWebToken(userID uuid.UUID) (string, time.Time, error) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       userID.String(),
		"username": "",
		"is_admin": false,
		"role":     "web",
		"exp":      expiresAt.Unix(),
	})
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

// InvalidateWebToken помечает веб-токен как недействительный (после выхода)
func (s *ServiceImpl) InvalidateWebToken(tokenString string) {
	s.addToInvalidTokenCache(tokenString)
}

// ValidateWebToken проверяет JWT веб-клиента и возвращает user_id (в claims.ID)
func (s *ServiceImpl) ValidateWebToken(tokenString string) (*models.TokenClaims, error) {
	if s.isInInvalidTokenCache(tokenString) {
		return nil, errors.New("токен недействителен")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("недействительный токен")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("недействительные данные токена")
	}
	role, _ := claims["role"].(string)
	if role != "web" {
		return nil, errors.New("токен не для веб-клиента")
	}
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, errors.New("токен истёк")
		}
	}
	idStr, _ := claims["id"].(string)
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	return &models.TokenClaims{
		ID:   userID,
		Role: "web",
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

const minPasswordLength = 8
const bcryptCostWeb = 12

// RegisterSendCode отправляет код подтверждения на email при регистрации
func (s *ServiceImpl) RegisterSendCode(ctx context.Context, email, password, passwordConfirm string) error {
	if s.webUserAuth == nil || s.webEmailCodeSender == nil {
		return ErrWebAuthNotConfigured
	}
	email = normalizeEmail(email)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("неверный формат email")
	}
	if password != passwordConfirm {
		return fmt.Errorf("пароли не совпадают")
	}
	if len(password) < minPasswordLength {
		return fmt.Errorf("пароль должен быть не менее %d символов", minPasswordLength)
	}
	_, err := s.webUserAuth.GetUserByEmail(ctx, email)
	if err == nil {
		return ErrUserAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.webEmailCodeSender.SendRegisterCode(ctx, email)
}

// VerifyEmailAndRegister проверяет код и создаёт пользователя
func (s *ServiceImpl) VerifyEmailAndRegister(ctx context.Context, email, code, password, passwordConfirm string) (*models.WebAuthResponse, error) {
	if s.webUserAuth == nil || s.webEmailCodeSender == nil {
		return nil, ErrWebAuthNotConfigured
	}
	email = normalizeEmail(email)
	if !emailRegex.MatchString(email) {
		return nil, fmt.Errorf("неверный формат email")
	}
	if password != passwordConfirm {
		return nil, fmt.Errorf("пароли не совпадают")
	}
	if len(password) < minPasswordLength {
		return nil, fmt.Errorf("пароль должен быть не менее %d символов", minPasswordLength)
	}
	if !s.webEmailCodeSender.VerifyRegisterCode(email, code) {
		return nil, ErrInvalidVerificationCode
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCostWeb)
	if err != nil {
		return nil, err
	}
	user, err := s.webUserAuth.CreateWebUser(ctx, email, string(hashedPassword))
	if err != nil {
		return nil, err
	}
	token, expiresAt, err := s.GenerateWebToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &models.WebAuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      userToWebUser(user),
	}, nil
}

// LoginWeb вход по email и паролю
func (s *ServiceImpl) LoginWeb(ctx context.Context, email, password string) (*models.WebAuthResponse, error) {
	if s.webUserAuth == nil {
		return nil, ErrWebAuthNotConfigured
	}
	email = normalizeEmail(email)
	user, err := s.webUserAuth.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	token, expiresAt, err := s.GenerateWebToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &models.WebAuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      userToWebUser(user),
	}, nil
}

// ChangePasswordSendCode отправляет код на email для смены пароля
func (s *ServiceImpl) ChangePasswordSendCode(ctx context.Context, email string) error {
	if s.webUserAuth == nil || s.webEmailCodeSender == nil {
		return ErrWebAuthNotConfigured
	}
	email = normalizeEmail(email)
	_, err := s.webUserAuth.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCredentials
		}
		return err
	}
	return s.webEmailCodeSender.SendPasswordChangeCode(ctx, email)
}

// ChangePasswordConfirm проверяет код и обновляет пароль
func (s *ServiceImpl) ChangePasswordConfirm(ctx context.Context, email, code, newPassword, newPasswordConfirm string) error {
	if s.webUserAuth == nil || s.webEmailCodeSender == nil {
		return ErrWebAuthNotConfigured
	}
	email = normalizeEmail(email)
	if newPassword != newPasswordConfirm {
		return fmt.Errorf("пароли не совпадают")
	}
	if len(newPassword) < minPasswordLength {
		return fmt.Errorf("пароль должен быть не менее %d символов", minPasswordLength)
	}
	if !s.webEmailCodeSender.VerifyPasswordChangeCode(email, code) {
		return ErrInvalidVerificationCode
	}
	user, err := s.webUserAuth.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCostWeb)
	if err != nil {
		return err
	}
	return s.webUserAuth.UpdatePassword(ctx, user.ID, string(hashedPassword))
}

func userToWebUser(u *userModels.User) models.WebUser {
	return models.WebUser{
		ID:               u.ID,
		Email:            u.Email,
		TelegramID:       u.TelegramID,
		CarNumber:        u.CarNumber,
		CarNumberCountry: u.CarNumberCountry,
	}
}
