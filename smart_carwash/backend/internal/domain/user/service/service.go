package service

import (
	"carwash_backend/internal/domain/user/models"
	"carwash_backend/internal/domain/user/repository"
	"carwash_backend/internal/utils"
	"context"

	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики пользователей
type Service interface {
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByCarNumber(ctx context.Context, carNumber string) (*models.User, error)
	UpdateCarNumber(ctx context.Context, req *models.UpdateCarNumberRequest) (*models.UpdateCarNumberResponse, error)
	UpdateEmail(ctx context.Context, req *models.UpdateEmailRequest) (*models.UpdateEmailResponse, error)

	// Административные методы
	AdminListUsers(ctx context.Context, req *models.AdminListUsersRequest) (*models.AdminListUsersResponse, error)
	AdminGetUser(ctx context.Context, req *models.AdminGetUserRequest) (*models.AdminGetUserResponse, error)
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
func (s *ServiceImpl) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Проверяем, существует ли пользователь
	existingUser, err := s.repo.GetUserByTelegramID(ctx, req.TelegramID)
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
	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByTelegramID получает пользователя по Telegram ID
func (s *ServiceImpl) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	return s.repo.GetUserByTelegramID(ctx, telegramID)
}

// GetUserByID получает пользователя по ID
func (s *ServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

// GetUserByCarNumber получает пользователя по номеру автомобиля
func (s *ServiceImpl) GetUserByCarNumber(ctx context.Context, carNumber string) (*models.User, error) {
	return s.repo.GetUserByCarNumber(ctx, carNumber)
}

// AdminListUsers получает список пользователей для администратора
func (s *ServiceImpl) AdminListUsers(ctx context.Context, req *models.AdminListUsersRequest) (*models.AdminListUsersResponse, error) {
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
	users, total, err := s.repo.GetUsersWithPagination(ctx, limit, offset)
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
func (s *ServiceImpl) AdminGetUser(ctx context.Context, req *models.AdminGetUserRequest) (*models.AdminGetUserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &models.AdminGetUserResponse{
		User: *user,
	}, nil
}

// UpdateCarNumber обновляет номер машины пользователя
func (s *ServiceImpl) UpdateCarNumber(ctx context.Context, req *models.UpdateCarNumberRequest) (*models.UpdateCarNumberResponse, error) {
	// Нормализация номера машины (без валидации)
	normalizedCarNumber := utils.NormalizeLicensePlate(req.CarNumber)

	// Устанавливаем дефолтную страну если не указана
	country := req.CarNumberCountry
	if country == "" {
		country = "RUS"
	}

	// Получаем пользователя
	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Обновляем номер машины и страну нормализованными значениями
	user.CarNumber = normalizedCarNumber
	user.CarNumberCountry = country
	err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.UpdateCarNumberResponse{
		Success: true,
		User:    *user,
	}, nil
}

// UpdateEmail обновляет email пользователя
func (s *ServiceImpl) UpdateEmail(ctx context.Context, req *models.UpdateEmailRequest) (*models.UpdateEmailResponse, error) {
	// Валидация email
	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRegex.MatchString(req.Email) {
		return nil, fmt.Errorf("неверный формат email")
	}

	// Получаем пользователя
	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Обновляем email
	user.Email = req.Email
	err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.UpdateEmailResponse{
		Success: true,
		User:    *user,
	}, nil
}
