package service

import (
	"carwash_backend/internal/domain/user/models"
	"carwash_backend/internal/domain/user/repository"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики пользователей
type Service interface {
	CreateUser(req *models.CreateUserRequest) (*models.User, error)
	GetUserByTelegramID(telegramID int64) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)

	// Административные методы
	AdminListUsers(req *models.AdminListUsersRequest) (*models.AdminListUsersResponse, error)
	AdminGetUser(req *models.AdminGetUserRequest) (*models.AdminGetUserResponse, error)
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

// GetUserByID получает пользователя по ID
func (s *ServiceImpl) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

// AdminListUsers получает список пользователей для администратора
func (s *ServiceImpl) AdminListUsers(req *models.AdminListUsersRequest) (*models.AdminListUsersResponse, error) {
	// Устанавливаем значения по умолчанию
	limit := 50
	offset := 0

	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	// Получаем пользователей с пагинацией
	users, total, err := s.repo.GetUsersWithPagination(limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.AdminListUsersResponse{
		Users:  users,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// AdminGetUser получает пользователя по ID для администратора
func (s *ServiceImpl) AdminGetUser(req *models.AdminGetUserRequest) (*models.AdminGetUserResponse, error) {
	user, err := s.repo.GetUserByID(req.ID)
	if err != nil {
		return nil, err
	}

	return &models.AdminGetUserResponse{
		User: *user,
	}, nil
}
