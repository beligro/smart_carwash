package service

import (
	"carwash_backend/internal/models"
	"carwash_backend/internal/repository"
)

// Service интерфейс для бизнес-логики
type Service interface {
	// Пользователи
	CreateUser(req *models.CreateUserRequest) (*models.User, error)
	GetUserByTelegramID(telegramID int64) (*models.User, error)

	// Информация о мойке
	GetWashInfo() (*models.WashInfo, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo repository.Repository
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository) *ServiceImpl {
	return &ServiceImpl{repo: repo}
}

// CreateUser создает нового пользователя
func (s *ServiceImpl) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Проверяем, существует ли пользователь
	existingUser, err := s.repo.GetUserByTelegramID(req.TelegramID)
	if err == nil {
		// Пользователь уже существует, возвращаем его
		return existingUser, nil
	}

	// Создаем нового пользователя
	user := &models.User{
		TelegramID: req.TelegramID,
		Username:   req.Username,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		IsAdmin:    false, // По умолчанию пользователь не админ
	}

	// Сохраняем пользователя в базе данных
	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByTelegramID получает пользователя по Telegram ID
func (s *ServiceImpl) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	return s.repo.GetUserByTelegramID(telegramID)
}

// GetWashInfo получает информацию о мойке
func (s *ServiceImpl) GetWashInfo() (*models.WashInfo, error) {
	// Получаем все боксы мойки
	boxes, err := s.repo.GetAllWashBoxes()
	if err != nil {
		return nil, err
	}

	// Формируем ответ
	return &models.WashInfo{
		Boxes: boxes,
	}, nil
}
