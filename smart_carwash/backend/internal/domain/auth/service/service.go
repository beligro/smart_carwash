package service

import (
	"carwash_backend/internal/logger"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/auth/models"
	"carwash_backend/internal/domain/auth/repository"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
)

// Service интерфейс для бизнес-логики авторизации
type Service interface {
	// Методы для авторизации
	LoginAdmin(username, password string) (*models.LoginResponse, error)
	LoginCashier(ctx context.Context, username, password string) (*models.LoginResponse, error)
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
}

// invalidTokenCacheEntry запись в кэше невалидных токенов
type invalidTokenCacheEntry struct {
	expiresAt time.Time
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo   repository.Repository
	config *config.Config
	// Кэш невалидных токенов: токен -> время истечения кэша
	// TTL = 5 минут - чтобы не забивать память навсегда
	invalidTokenCache sync.Map
	cacheTTL          time.Duration
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, config *config.Config) *ServiceImpl {
	service := &ServiceImpl{
		repo:     repo,
		config:   config,
		cacheTTL: 5 * time.Minute, // TTL для кэша невалидных токенов
	}

	// Запускаем фоновую очистку кэша
	go service.cleanupInvalidTokenCache()

	return service
}

// LoginAdmin авторизует администратора
func (s *ServiceImpl) LoginAdmin(username, password string) (*models.LoginResponse, error) {
	// Проверяем, совпадают ли учетные данные с данными администратора из конфигурации
	if username != s.config.AdminUsername || password != s.config.AdminPassword {
		return nil, ErrInvalidCredentials
	}

	// Создаем JWT токен для администратора
	claims := models.TokenClaims{
		ID:       uuid.New(), // Для администратора генерируем случайный ID
		Username: username,
		IsAdmin:  true,
	}

	// Генерируем токен
	token, expiresAt, err := s.generateToken(claims)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		IsAdmin:   true,
	}, nil
}

// LoginCashier авторизует кассира
func (s *ServiceImpl) LoginCashier(ctx context.Context, username, password string) (*models.LoginResponse, error) {
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

	return &models.TokenClaims{
		ID:       id,
		Username: username,
		IsAdmin:  isAdmin,
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
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // Смена длится 24 часа

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
		"id":       claims.ID.String(),
		"username": claims.Username,
		"is_admin": claims.IsAdmin,
		"exp":      expiresAt.Unix(),
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
