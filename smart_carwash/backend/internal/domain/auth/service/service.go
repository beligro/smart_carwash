package service

import (
	"errors"
	"fmt"
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

	// ErrAdminAlreadyExists возвращается при попытке создать второго администратора
	ErrAdminAlreadyExists = errors.New("администратор уже существует")
)

// Service интерфейс для бизнес-логики авторизации
type Service interface {
	// Методы для авторизации
	LoginAdmin(username, password string) (*models.LoginResponse, error)
	LoginCashier(username, password string) (*models.LoginResponse, error)
	ValidateToken(token string) (*models.TokenClaims, error)
	Logout(token string) error

	// Методы для управления кассирами
	CreateCashier(req *models.CreateCashierRequest) (*models.CreateCashierResponse, error)
	UpdateCashier(req *models.UpdateCashierRequest) (*models.UpdateCashierResponse, error)
	DeleteCashier(id uuid.UUID) error
	GetCashiers() (*models.GetCashiersResponse, error)
	GetCashierByID(id uuid.UUID) (*models.Cashier, error)

	// Методы для двухфакторной аутентификации (заглушки для будущей реализации)
	EnableTwoFactorAuth(userID uuid.UUID) (string, error)
	DisableTwoFactorAuth(userID uuid.UUID) error
	VerifyTwoFactorCode(userID uuid.UUID, code string) (bool, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo   repository.Repository
	config *config.Config
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, config *config.Config) *ServiceImpl {
	return &ServiceImpl{
		repo:   repo,
		config: config,
	}
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
func (s *ServiceImpl) LoginCashier(username, password string) (*models.LoginResponse, error) {
	// Получаем кассира по имени пользователя
	cashier, err := s.repo.GetCashierByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrCashierNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Проверяем, активен ли кассир
	if !cashier.IsActive {
		return nil, ErrCashierInactive
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(cashier.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

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
	if err := s.repo.CreateCashierSession(session); err != nil {
		if errors.Is(err, repository.ErrActiveCashierSessionExists) {
			return nil, ErrAnotherCashierActive
		}
		return nil, err
	}

	// Обновляем время последнего входа кассира
	now := time.Now()
	cashier.LastLogin = &now
	if err := s.repo.UpdateCashier(cashier); err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		IsAdmin:   false,
	}, nil
}

// ValidateToken проверяет JWT токен и возвращает данные пользователя
func (s *ServiceImpl) ValidateToken(tokenString string) (*models.TokenClaims, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена
	if !token.Valid {
		return nil, errors.New("недействительный токен")
	}

	// Получаем данные из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("недействительные данные токена")
	}

	// Проверяем, не истек ли токен
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, errors.New("токен истек")
		}
	}

	// Если это токен кассира, проверяем, существует ли сессия
	isAdmin, _ := claims["is_admin"].(bool)
	if !isAdmin {
		// Получаем сессию по токену
		session, err := s.repo.GetCashierSessionByToken(tokenString)
		if err != nil {
			return nil, err
		}
		if session == nil {
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

// Logout завершает сессию пользователя
func (s *ServiceImpl) Logout(token string) error {
	// Проверяем токен
	claims, err := s.ValidateToken(token)
	if err != nil {
		return err
	}

	// Если это администратор, просто возвращаем успех
	if claims.IsAdmin {
		return nil
	}

	// Если это кассир, удаляем сессию
	session, err := s.repo.GetCashierSessionByToken(token)
	if err != nil {
		return err
	}
	if session == nil {
		return nil // Сессия уже удалена или истекла
	}

	return s.repo.DeleteCashierSession(session.ID)
}

// CreateCashier создает нового кассира
func (s *ServiceImpl) CreateCashier(req *models.CreateCashierRequest) (*models.CreateCashierResponse, error) {
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
	if err := s.repo.CreateCashier(cashier); err != nil {
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
func (s *ServiceImpl) UpdateCashier(req *models.UpdateCashierRequest) (*models.UpdateCashierResponse, error) {
	// Получаем кассира по ID
	cashier, err := s.repo.GetCashierByID(req.ID)
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
	if err := s.repo.UpdateCashier(cashier); err != nil {
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
func (s *ServiceImpl) DeleteCashier(id uuid.UUID) error {
	// Проверяем, существует ли кассир
	_, err := s.repo.GetCashierByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCashierNotFound) {
			return fmt.Errorf("кассир с ID %s не найден", id)
		}
		return err
	}

	// Удаляем все сессии кассира
	if err := s.repo.DeleteCashierSession(id); err != nil {
		return err
	}

	// Удаляем кассира
	return s.repo.DeleteCashier(id)
}

// GetCashiers возвращает список всех кассиров
func (s *ServiceImpl) GetCashiers() (*models.GetCashiersResponse, error) {
	cashiers, err := s.repo.ListCashiers()
	if err != nil {
		return nil, err
	}

	return &models.GetCashiersResponse{
		Cashiers: cashiers,
	}, nil
}

// GetCashierByID возвращает кассира по ID
func (s *ServiceImpl) GetCashierByID(id uuid.UUID) (*models.Cashier, error) {
	return s.repo.GetCashierByID(id)
}

// EnableTwoFactorAuth включает двухфакторную аутентификацию для пользователя
// (заглушка для будущей реализации)
func (s *ServiceImpl) EnableTwoFactorAuth(userID uuid.UUID) (string, error) {
	// Здесь будет логика генерации секрета и QR-кода для двухфакторной аутентификации
	// Пока просто заглушка
	settings := &models.TwoFactorAuthSettings{
		UserID:    userID,
		IsEnabled: true,
		Secret:    "dummy_secret", // В реальной реализации здесь будет настоящий секрет
	}

	if err := s.repo.SaveTwoFactorAuthSettings(settings); err != nil {
		return "", err
	}

	return "dummy_qr_code", nil // В реальной реализации здесь будет URL для QR-кода
}

// DisableTwoFactorAuth отключает двухфакторную аутентификацию для пользователя
// (заглушка для будущей реализации)
func (s *ServiceImpl) DisableTwoFactorAuth(userID uuid.UUID) error {
	settings := &models.TwoFactorAuthSettings{
		UserID:    userID,
		IsEnabled: false,
		Secret:    "",
	}

	return s.repo.SaveTwoFactorAuthSettings(settings)
}

// VerifyTwoFactorCode проверяет код двухфакторной аутентификации
// (заглушка для будущей реализации)
func (s *ServiceImpl) VerifyTwoFactorCode(userID uuid.UUID, code string) (bool, error) {
	// Здесь будет логика проверки кода двухфакторной аутентификации
	// Пока просто заглушка
	settings, err := s.repo.GetTwoFactorAuthSettings(userID)
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
