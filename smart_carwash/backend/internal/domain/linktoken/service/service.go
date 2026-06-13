package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/linktoken/models"
	"carwash_backend/internal/domain/linktoken/repository"
	userRepo "carwash_backend/internal/domain/user/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const tokenLength = 12
const tokenTTL = 15 * time.Minute

// Service сервис привязки Telegram через link_tokens
type Service interface {
	CreateToken(ctx context.Context, userID uuid.UUID) (*models.CreateLinkTokenResponse, error)
	ConsumeToken(ctx context.Context, token string, telegramID int64, telegramFirstName, telegramLastName, telegramUserName string) error
}

// ServiceImpl реализация
type ServiceImpl struct {
	repo        repository.Repository
	userRepo    userRepo.Repository
	sessionRepo SessionRepository
	cfg         *config.Config
}

// SessionRepository минимальный интерфейс для переноса сессий
type SessionRepository interface {
	ReassignSessionsToUser(ctx context.Context, fromUserID, toUserID uuid.UUID) error
}

// NewService создает сервис
func NewService(repo repository.Repository, userRepo userRepo.Repository, sessionRepo SessionRepository, cfg *config.Config) *ServiceImpl {
	return &ServiceImpl{
		repo:         repo,
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		cfg:          cfg,
	}
}

// CreateToken создает одноразовый токен и возвращает ссылку для привязки
func (s *ServiceImpl) CreateToken(ctx context.Context, userID uuid.UUID) (*models.CreateLinkTokenResponse, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	lt := &models.LinkToken{
		Token:     "link_" + token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(tokenTTL),
	}
	if err := s.repo.Create(ctx, lt); err != nil {
		return nil, err
	}

	botUsername := s.cfg.TelegramUsername
	if botUsername == "" {
		botUsername = "your_bot"
	}
	link := fmt.Sprintf("https://t.me/%s?start=link_%s", botUsername, token)
	return &models.CreateLinkTokenResponse{Link: link}, nil
}

// ConsumeToken обрабатывает переход по ссылке привязки: привязывает telegram_id к веб-пользователю и переносит данные из Telegram-аккаунта (имя, фамилия, никнейм, машина, email и т.д.)
func (s *ServiceImpl) ConsumeToken(ctx context.Context, token string, telegramID int64, telegramFirstName, telegramLastName, telegramUserName string) error {
	fullToken := token
	if len(token) > 5 && token[:5] != "link_" {
		fullToken = "link_" + token
	}

	lt, err := s.repo.GetByToken(ctx, fullToken)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("токен не найден или истёк")
		}
		return err
	}
	if lt.UsedAt != nil {
		return fmt.Errorf("токен уже использован")
	}
	if time.Now().After(lt.ExpiresAt) {
		return fmt.Errorf("токен истёк")
	}

	webUser, err := s.userRepo.GetUserByID(ctx, lt.UserID)
	if err != nil {
		return err
	}

	existingUser, _ := s.userRepo.GetUserByTelegramID(ctx, telegramID)
	if existingUser != nil && existingUser.ID != webUser.ID {
		// Конфликт: уже есть другой пользователь с этим telegram_id. Переносим его сессии на веб-пользователя.
		if err := s.sessionRepo.ReassignSessionsToUser(ctx, existingUser.ID, webUser.ID); err != nil {
			return fmt.Errorf("ошибка объединения сессий: %w", err)
		}
		// Переносим данные из Telegram-пользователя в веб-пользователя.
		// ВАЖНО: не затираем уже существующие критичные поля веб-пользователя (например, email).
		if existingUser.Username != "" {
			webUser.Username = existingUser.Username
		}
		if existingUser.FirstName != "" {
			webUser.FirstName = existingUser.FirstName
		}
		if existingUser.LastName != "" {
			webUser.LastName = existingUser.LastName
		}
		if existingUser.CarNumber != "" {
			webUser.CarNumber = existingUser.CarNumber
		}
		if existingUser.CarNumberCountry != "" {
			webUser.CarNumberCountry = existingUser.CarNumberCountry
		}
		// Обнуляем telegram_id у старого пользователя
		existingUser.TelegramID = nil
		if err := s.userRepo.UpdateUser(ctx, existingUser); err != nil {
			return err
		}
	}

	// Обновляем веб-пользователя: telegram_id и данные из Telegram (если не перенесли из existingUser)
	webUser.TelegramID = &telegramID
	webUser.FirstName = telegramFirstName
	webUser.LastName = telegramLastName
	webUser.Username = telegramUserName
	if existingUser != nil {
		// Уже скопировали выше CarNumber, CarNumberCountry, Email
	} else {
		// Нет старого Telegram-пользователя — только имя/фамилия/ник из текущего сообщения (уже заданы выше)
	}

	if err := s.userRepo.UpdateUser(ctx, webUser); err != nil {
		return err
	}

	now := time.Now()
	return s.repo.MarkUsed(ctx, fullToken, now)
}

func generateToken() (string, error) {
	b := make([]byte, tokenLength/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:tokenLength], nil
}
