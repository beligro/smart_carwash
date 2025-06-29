package repository

import (
	"carwash_backend/internal/domain/session/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с сессиями в базе данных
type Repository interface {
	CreateSession(session *models.Session) error
	GetSessionByID(id uuid.UUID) (*models.Session, error)
	GetActiveSessionByUserID(userID uuid.UUID) (*models.Session, error)
	GetSessionByIdempotencyKey(key string) (*models.Session, error)
	UpdateSession(session *models.Session) error
	GetSessionsByStatus(status string) ([]models.Session, error)
	CountSessionsByStatus(status string) (int, error)
	GetUserSessionHistory(userID uuid.UUID, limit, offset int) ([]models.Session, error)

	// Административные методы
	GetSessionsWithFilters(userID *uuid.UUID, boxID *uuid.UUID, boxNumber *int, status *string, serviceType *string, dateFrom *time.Time, dateTo *time.Time, limit int, offset int) ([]models.Session, int, error)
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateSession создает новую сессию мойки
func (r *PostgresRepository) CreateSession(session *models.Session) error {
	return r.db.Create(session).Error
}

// GetSessionByID получает сессию по ID
func (r *PostgresRepository) GetSessionByID(id uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.First(&session, id).Error
	if err != nil {
		return nil, err
	}

	// Если у сессии есть BoxID, получаем номер бокса
	if session.BoxID != nil {
		var boxNumber int
		err = r.db.Table("wash_boxes").Where("id = ?", *session.BoxID).Select("number").Scan(&boxNumber).Error
		if err == nil {
			session.BoxNumber = &boxNumber
		}
	}

	return &session, nil
}

// GetActiveSessionByUserID получает активную сессию пользователя
func (r *PostgresRepository) GetActiveSessionByUserID(userID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("user_id = ? AND status IN (?, ?, ?)",
		userID,
		models.SessionStatusCreated,
		models.SessionStatusAssigned,
		models.SessionStatusActive).
		Order("created_at DESC").
		First(&session).Error
	if err != nil {
		return nil, err
	}

	// Если у сессии есть BoxID, получаем номер бокса
	if session.BoxID != nil {
		var boxNumber int
		err = r.db.Table("wash_boxes").Where("id = ?", *session.BoxID).Select("number").Scan(&boxNumber).Error
		if err == nil {
			session.BoxNumber = &boxNumber
		}
	}

	return &session, nil
}

// GetSessionByIdempotencyKey получает сессию по ключу идемпотентности
func (r *PostgresRepository) GetSessionByIdempotencyKey(key string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("idempotency_key = ?", key).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession обновляет сессию
func (r *PostgresRepository) UpdateSession(session *models.Session) error {
	return r.db.Save(session).Error
}

// GetSessionsByStatus получает сессии по статусу
func (r *PostgresRepository) GetSessionsByStatus(status string) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.Where("status = ?", status).Order("created_at ASC").Find(&sessions).Error
	return sessions, err
}

// CountSessionsByStatus подсчитывает количество сессий с определенным статусом
func (r *PostgresRepository) CountSessionsByStatus(status string) (int, error) {
	var count int64
	err := r.db.Model(&models.Session{}).Where("status = ?", status).Count(&count).Error
	return int(count), err
}

// GetUserSessionHistory получает историю сессий пользователя
func (r *PostgresRepository) GetUserSessionHistory(userID uuid.UUID, limit, offset int) ([]models.Session, error) {
	var sessions []models.Session
	query := r.db.Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&sessions).Error
	return sessions, err
}

// GetSessionsWithFilters получает сессии с фильтрацией для администратора
func (r *PostgresRepository) GetSessionsWithFilters(userID *uuid.UUID, boxID *uuid.UUID, boxNumber *int, status *string, serviceType *string, dateFrom *time.Time, dateTo *time.Time, limit int, offset int) ([]models.Session, int, error) {
	var sessions []models.Session
	var total int64

	query := r.db.Model(&models.Session{})

	// Применяем фильтры
	if userID != nil {
		query = query.Where("sessions.user_id = ?", *userID)
	}
	if boxID != nil {
		query = query.Where("sessions.box_id = ?", *boxID)
	}
	if boxNumber != nil {
		// Используем JOIN для фильтрации по номеру бокса
		query = query.Joins("JOIN wash_boxes ON sessions.box_id = wash_boxes.id").
			Where("wash_boxes.number = ?", *boxNumber)
	}
	if status != nil {
		query = query.Where("sessions.status = ?", *status)
	}
	if serviceType != nil {
		query = query.Where("sessions.service_type = ?", *serviceType)
	}
	if dateFrom != nil {
		query = query.Where("sessions.created_at >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("sessions.created_at <= ?", *dateTo)
	}

	// Получаем общее количество
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией и сортировкой
	err = query.Order("sessions.created_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error
	if err != nil {
		return nil, 0, err
	}

	// Для каждой сессии получаем номер бокса, если он есть
	for i := range sessions {
		if sessions[i].BoxID != nil {
			var boxNumber int
			err = r.db.Table("wash_boxes").Where("id = ?", *sessions[i].BoxID).Select("number").Scan(&boxNumber).Error
			if err == nil {
				sessions[i].BoxNumber = &boxNumber
			}
		}
	}

	return sessions, int(total), nil
}
