package repository

import (
	"carwash_backend/internal/domain/session/models"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с сессиями в базе данных
type Repository interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetActiveSessionByUserID(ctx context.Context, userID uuid.UUID) (*models.Session, error)
	GetActiveSessionByCarNumber(ctx context.Context, carNumber string) (*models.Session, error)
	GetUserSessionForPayment(ctx context.Context, userID uuid.UUID) (*models.Session, error)
	GetSessionByIdempotencyKey(ctx context.Context, key string) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	UpdateSessionFields(ctx context.Context, sessionID uuid.UUID, fields map[string]interface{}) error
	GetSessionsByStatus(ctx context.Context, status string) ([]models.Session, error)
	CountSessionsByStatus(ctx context.Context, status string) (int, error)
	GetUserSessionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Session, error)

	// Административные методы
	GetSessionsWithFilters(ctx context.Context, userID *uuid.UUID, boxID *uuid.UUID, boxNumber *int, status *string, serviceType *string, dateFrom *time.Time, dateTo *time.Time, limit int, offset int) ([]models.Session, int, error)

	// Метод для проверки завершенных сессий между уборками
	GetCompletedSessionsBetween(ctx context.Context, boxID uuid.UUID, dateFrom, dateTo time.Time) (int, error)
	GetCompletedSessionsCreatedAtSinceForBoxes(ctx context.Context, boxIDs []uuid.UUID, since time.Time) ([]struct {
		BoxID     uuid.UUID
		CreatedAt time.Time
	}, error)

	// Методы с блокировкой для предотвращения дедлоков
	GetSessionByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Session, error)
	UpdateSessionInTransaction(ctx context.Context, session *models.Session) error
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// CreateSession создает новую сессию мойки
func (r *PostgresRepository) CreateSession(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSessionByID получает сессию по ID
func (r *PostgresRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).First(&session, id).Error
	if err != nil {
		return nil, err
	}

	// Если у сессии есть BoxID, получаем номер бокса
	if session.BoxID != nil {
		var boxNumber int
		err = r.db.WithContext(ctx).Table("wash_boxes").Where("id = ?", *session.BoxID).Select("number").Scan(&boxNumber).Error
		if err == nil {
			session.BoxNumber = &boxNumber
		}
	}

	return &session, nil
}

// GetActiveSessionByUserID получает активную сессию пользователя
func (r *PostgresRepository) GetActiveSessionByUserID(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("user_id = ? AND status IN (?, ?, ?, ?)",
		userID,
		models.SessionStatusCreated,
		models.SessionStatusInQueue,
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
		err = r.db.WithContext(ctx).Table("wash_boxes").Where("id = ?", *session.BoxID).Select("number").Scan(&boxNumber).Error
		if err == nil {
			session.BoxNumber = &boxNumber
		}
	}

	return &session, nil
}

// GetUserSessionForPayment получает сессию пользователя для PaymentPage (включая payment_failed)
func (r *PostgresRepository) GetUserSessionForPayment(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("user_id = ? AND status IN (?, ?, ?, ?, ?)",
		userID,
		models.SessionStatusCreated,
		models.SessionStatusInQueue,
		models.SessionStatusAssigned,
		models.SessionStatusActive,
		models.SessionStatusPaymentFailed).
		Order("created_at DESC").
		First(&session).Error
	if err != nil {
		return nil, err
	}

	// Если у сессии есть BoxID, получаем номер бокса
	if session.BoxID != nil {
		var boxNumber int
		err = r.db.WithContext(ctx).Table("wash_boxes").Where("id = ?", *session.BoxID).Select("number").Scan(&boxNumber).Error
		if err == nil {
			session.BoxNumber = &boxNumber
		}
	}

	return &session, nil
}

// CheckActiveSessionWithLock проверяет активную сессию пользователя с учетом временной блокировки
func (r *PostgresRepository) CheckActiveSessionWithLock(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("user_id = ? AND status IN (?, ?, ?, ?) AND created_at > ?",
		userID,
		models.SessionStatusCreated,
		models.SessionStatusInQueue,
		models.SessionStatusAssigned,
		models.SessionStatusActive,
		time.Now().Add(-30*time.Second)).
		Order("created_at DESC").
		First(&session).Error
	if err != nil {
		return nil, err
	}

	// Если у сессии есть BoxID, получаем номер бокса
	if session.BoxID != nil {
		var boxNumber int
		err = r.db.WithContext(ctx).Table("wash_boxes").Where("id = ?", *session.BoxID).Select("number").Scan(&boxNumber).Error
		if err == nil {
			session.BoxNumber = &boxNumber
		}
	}

	return &session, nil
}

// GetSessionByIdempotencyKey получает сессию по ключу идемпотентности
func (r *PostgresRepository) GetSessionByIdempotencyKey(ctx context.Context, key string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession обновляет сессию
func (r *PostgresRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	// Используем Save для обновления всех полей
	// ВАЖНО: Save обновляет все поля, включая updated_at (автоматически через GORM)
	// но НЕ обновляет status_updated_at если мы его явно не изменили
	return r.db.WithContext(ctx).Save(session).Error
}

// UpdateSessionFields обновляет только указанные поля сессии
func (r *PostgresRepository) UpdateSessionFields(ctx context.Context, sessionID uuid.UUID, fields map[string]interface{}) error {
	// Используем Updates для обновления только указанных полей
	// Это безопаснее чем Save, так как не затрагивает другие поля
	return r.db.WithContext(ctx).Model(&models.Session{}).Where("id = ?", sessionID).Updates(fields).Error
}

// GetSessionsByStatus получает сессии по статусу
func (r *PostgresRepository) GetSessionsByStatus(ctx context.Context, status string) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at ASC").Find(&sessions).Error
	return sessions, err
}

// CountSessionsByStatus подсчитывает количество сессий с определенным статусом
func (r *PostgresRepository) CountSessionsByStatus(ctx context.Context, status string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Session{}).Where("status = ?", status).Count(&count).Error
	return int(count), err
}

// GetUserSessionHistory получает историю сессий пользователя
func (r *PostgresRepository) GetUserSessionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Session, error) {
	var sessions []models.Session
	db := r.db.WithContext(ctx)
	query := db.Where("user_id = ?", userID).
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
func (r *PostgresRepository) GetSessionsWithFilters(ctx context.Context, userID *uuid.UUID, boxID *uuid.UUID, boxNumber *int, status *string, serviceType *string, dateFrom *time.Time, dateTo *time.Time, limit int, offset int) ([]models.Session, int, error) {
	var sessions []models.Session

	// Создаем функцию для применения фильтров к запросу
	applyFilters := func(query *gorm.DB) *gorm.DB {
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
		return query
	}

	// Подсчитываем общее количество сессий с фильтрами (без пагинации)
	var total int64
	countQuery := r.db.WithContext(ctx).Model(&models.Session{})
	countQuery = applyFilters(countQuery)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией и сортировкой
	err := func() error {
		query := r.db.WithContext(ctx).Model(&models.Session{})
		query = applyFilters(query)
		return query.Order("sessions.created_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error
	}()
	if err != nil {
		return nil, 0, err
	}

	// Для всех сессий одним запросом получаем номера боксов и мапим по id
	// Сначала собираем уникальные box_id
	boxIDSet := make(map[uuid.UUID]struct{})
	for i := range sessions {
		if sessions[i].BoxID != nil {
			boxIDSet[*sessions[i].BoxID] = struct{}{}
		}
	}

	if len(boxIDSet) > 0 {
		boxIDs := make([]uuid.UUID, 0, len(boxIDSet))
		for id := range boxIDSet {
			boxIDs = append(boxIDs, id)
		}

		type boxRow struct {
			ID     uuid.UUID
			Number int
		}
		var rows []boxRow
		if err := r.db.WithContext(ctx).Table("wash_boxes").
			Select("id, number").
			Where("id IN ?", boxIDs).
			Find(&rows).Error; err == nil {
			numberByID := make(map[uuid.UUID]int, len(rows))
			for _, row := range rows {
				numberByID[row.ID] = row.Number
			}
			for i := range sessions {
				if sessions[i].BoxID != nil {
					if num, ok := numberByID[*sessions[i].BoxID]; ok {
						n := num
						sessions[i].BoxNumber = &n
					}
				}
			}
		}
	}

	return sessions, int(total), nil
}

// GetCompletedSessionsBetween получает количество завершенных сессий для бокса между указанными датами
func (r *PostgresRepository) GetCompletedSessionsBetween(ctx context.Context, boxID uuid.UUID, dateFrom, dateTo time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("box_id = ? AND status = ? AND created_at >= ? AND created_at <= ?",
			boxID, models.SessionStatusComplete, dateFrom, dateTo).
		Count(&count).Error
	return int(count), err
}

// GetCompletedSessionsCreatedAtSinceForBoxes возвращает пары (box_id, created_at) для завершенных сессий
// для множества боксов начиная с указанной даты (общей нижней границы)
func (r *PostgresRepository) GetCompletedSessionsCreatedAtSinceForBoxes(ctx context.Context, boxIDs []uuid.UUID, since time.Time) ([]struct {
	BoxID     uuid.UUID
	CreatedAt time.Time
}, error) {
	if len(boxIDs) == 0 {
		return nil, nil
	}
	var rows []struct {
		BoxID     uuid.UUID
		CreatedAt time.Time
	}
	err := r.db.WithContext(ctx).Model(&models.Session{}).
		Select("box_id, created_at").
		Where("status = ? AND box_id IN ? AND created_at >= ?", models.SessionStatusComplete, boxIDs, since).
		Find(&rows).Error
	return rows, err
}

// GetSessionByIDForUpdate получает сессию по ID (без блокировки)
func (r *PostgresRepository) GetSessionByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).First(&session, id).Error
	if err != nil {
		return nil, err
	}

	// Если у сессии есть BoxID, получаем номер бокса
	if session.BoxID != nil {
		var boxNumber int
		err = r.db.WithContext(ctx).Table("wash_boxes").Where("id = ?", *session.BoxID).Select("number").Scan(&boxNumber).Error
		if err == nil {
			session.BoxNumber = &boxNumber
		}
	}

	return &session, nil
}

// UpdateSessionInTransaction обновляет сессию в рамках транзакции
// Этот метод должен вызываться внутри транзакции
func (r *PostgresRepository) UpdateSessionInTransaction(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// GetActiveSessionByCarNumber получает активную сессию по номеру автомобиля
func (r *PostgresRepository) GetActiveSessionByCarNumber(ctx context.Context, carNumber string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("car_number = ? AND status IN (?, ?, ?, ?)",
		carNumber,
		models.SessionStatusCreated,
		models.SessionStatusInQueue,
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
		err = r.db.WithContext(ctx).Table("wash_boxes").Where("id = ?", *session.BoxID).Select("number").Scan(&boxNumber).Error
		if err == nil {
			session.BoxNumber = &boxNumber
		}
	}

	return &session, nil
}
