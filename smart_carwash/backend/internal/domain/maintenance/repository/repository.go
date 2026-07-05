package repository

import (
	"context"
	"time"

	"carwash_backend/internal/domain/maintenance/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс доступа к данным сервисных нарядов.
type Repository interface {
	// Справочники
	GetSymptomsByBoxType(ctx context.Context, boxType string) ([]models.SymptomType, error)
	GetSymptomByID(ctx context.Context, id int) (*models.SymptomType, error)
	GetCatalog(ctx context.Context, boxType string) ([]models.CatalogGroup, error)
	GetComponentCarrier(ctx context.Context, componentID int) (string, error)

	// Наряды
	CreateTicket(ctx context.Context, ticket *models.ServiceTicket) error
	GetOpenTicketByBoxNumber(ctx context.Context, boxNumber int) (*models.ServiceTicket, error)
	GetAllOpenTickets(ctx context.Context) ([]models.ServiceTicket, error)
	GetTicketByID(ctx context.Context, id uuid.UUID) (*models.ServiceTicket, error)
	CloseTicket(ctx context.Context, ticket *models.ServiceTicket, works []models.TicketWork) error
	ListTickets(ctx context.Context, status *string, boxNumber *int, since *time.Time, limit int) ([]models.TicketView, error)

	// Получатели уведомлений
	ActiveRecipients(ctx context.Context, eventType string) ([]models.NotificationRecipient, error)
	ListRecipients(ctx context.Context) ([]models.NotificationRecipient, error)
	CreateRecipient(ctx context.Context, r *models.NotificationRecipient) error
	SetRecipientActive(ctx context.Context, id uuid.UUID, active bool) error

	// Вспомогательное
	GetCashierUsername(ctx context.Context, cashierID uuid.UUID) (string, error)
	ComputeBoxMotorHours(ctx context.Context, boxNumber int) (int, error)
}

// PostgresRepository реализация Repository на GORM/Postgres.
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создаёт репозиторий.
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetSymptomsByBoxType возвращает активные симптомы для типа бокса
// плюс общие симптомы (box_type='any', напр. «Бокс заблокирован»).
func (r *PostgresRepository) GetSymptomsByBoxType(ctx context.Context, boxType string) ([]models.SymptomType, error) {
	var rows []models.SymptomType
	err := r.db.WithContext(ctx).
		Where("is_active = true AND (box_type = ? OR box_type = ?)", boxType, models.BoxTypeAny).
		Order("sort_order, id").
		Find(&rows).Error
	return rows, err
}

// GetSymptomByID возвращает симптом по id (или nil).
func (r *PostgresRepository) GetSymptomByID(ctx context.Context, id int) (*models.SymptomType, error) {
	var s models.SymptomType
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// GetCatalog возвращает группы работ (с деталями), подходящие для типа бокса (+ группы 'any').
func (r *PostgresRepository) GetCatalog(ctx context.Context, boxType string) ([]models.CatalogGroup, error) {
	var groups []models.ComponentGroup
	if err := r.db.WithContext(ctx).
		Where("box_type = ? OR box_type = ?", boxType, models.BoxTypeAny).
		Order("sort_order, id").
		Find(&groups).Error; err != nil {
		return nil, err
	}

	var types []models.ComponentType
	if err := r.db.WithContext(ctx).
		Order("sort_order, id").
		Find(&types).Error; err != nil {
		return nil, err
	}

	byGroup := make(map[int][]models.ComponentType, len(types))
	for _, t := range types {
		byGroup[t.GroupID] = append(byGroup[t.GroupID], t)
	}

	result := make([]models.CatalogGroup, 0, len(groups))
	for _, g := range groups {
		result = append(result, models.CatalogGroup{
			ID:       g.ID,
			Name:     g.Name,
			Carrier:  g.Carrier,
			BoxType:  g.BoxType,
			Children: byGroup[g.ID],
		})
	}
	return result, nil
}

// GetComponentCarrier возвращает носитель (box|machine|any) группы указанной детали.
func (r *PostgresRepository) GetComponentCarrier(ctx context.Context, componentID int) (string, error) {
	var carrier string
	err := r.db.WithContext(ctx).
		Raw(`SELECT g.carrier FROM component_types t JOIN component_groups g ON g.id = t.group_id WHERE t.id = ?`, componentID).
		Scan(&carrier).Error
	return carrier, err
}

// CreateTicket создаёт наряд.
func (r *PostgresRepository) CreateTicket(ctx context.Context, ticket *models.ServiceTicket) error {
	if ticket.ID == uuid.Nil {
		ticket.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(ticket).Error
}

// GetOpenTicketByBoxNumber возвращает открытый наряд по номеру бокса (или nil).
func (r *PostgresRepository) GetOpenTicketByBoxNumber(ctx context.Context, boxNumber int) (*models.ServiceTicket, error) {
	var t models.ServiceTicket
	err := r.db.WithContext(ctx).
		Where("box_number = ? AND status = ?", boxNumber, models.TicketStatusOpen).
		Order("opened_at DESC").
		First(&t).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// GetAllOpenTickets возвращает все открытые наряды.
func (r *PostgresRepository) GetAllOpenTickets(ctx context.Context) ([]models.ServiceTicket, error) {
	var rows []models.ServiceTicket
	err := r.db.WithContext(ctx).Where("status = ?", models.TicketStatusOpen).Find(&rows).Error
	return rows, err
}

// GetTicketByID возвращает наряд по id.
func (r *PostgresRepository) GetTicketByID(ctx context.Context, id uuid.UUID) (*models.ServiceTicket, error) {
	var t models.ServiceTicket
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CloseTicket закрывает наряд и записывает работы (в транзакции).
func (r *PostgresRepository) CloseTicket(ctx context.Context, ticket *models.ServiceTicket, works []models.TicketWork) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ServiceTicket{}).
			Where("id = ?", ticket.ID).
			Updates(map[string]interface{}{
				"status":         models.TicketStatusClosed,
				"closed_at":      ticket.ClosedAt,
				"closed_by":      ticket.ClosedBy,
				"master_comment": ticket.MasterComment,
				"updated_at":     time.Now(),
			}).Error; err != nil {
			return err
		}
		for i := range works {
			if works[i].ID == uuid.Nil {
				works[i].ID = uuid.New()
			}
			if err := tx.Create(&works[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListTickets возвращает журнал нарядов с расшифровкой симптома и работами.
func (r *PostgresRepository) ListTickets(ctx context.Context, status *string, boxNumber *int, since *time.Time, limit int) ([]models.TicketView, error) {
	type row struct {
		models.ServiceTicket
		SymptomName  string
		SymptomGroup string
	}
	q := r.db.WithContext(ctx).
		Table("service_tickets st").
		Select(`st.*, s.name AS symptom_name, s.group_name AS symptom_group`).
		Joins(`LEFT JOIN symptom_types s ON s.id = st.symptom_id`)

	if status != nil && *status != "" {
		q = q.Where("st.status = ?", *status)
	}
	if boxNumber != nil {
		q = q.Where("st.box_number = ?", *boxNumber)
	}
	if since != nil {
		q = q.Where("st.opened_at >= ?", *since)
	}
	// Открытые наряды — сверху, затем по времени открытия убыванием.
	q = q.Order("CASE WHEN st.status = 'open' THEN 0 ELSE 1 END, st.opened_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []models.TicketView{}, nil
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, r0 := range rows {
		ids = append(ids, r0.ID)
	}

	// Работы по нарядам
	type workRow struct {
		TicketID      uuid.UUID
		ComponentID   int
		ComponentName string
		GroupName     string
		Action        string
		Carrier       string
		MotorHours    *int
	}
	var workRows []workRow
	if err := r.db.WithContext(ctx).
		Table("ticket_works tw").
		Select(`tw.ticket_id, tw.component_id, ct.name AS component_name, cg.name AS group_name, tw.action, tw.carrier, tw.motor_hours`).
		Joins(`LEFT JOIN component_types ct ON ct.id = tw.component_id`).
		Joins(`LEFT JOIN component_groups cg ON cg.id = ct.group_id`).
		Where("tw.ticket_id IN ?", ids).
		Order("tw.created_at").
		Scan(&workRows).Error; err != nil {
		return nil, err
	}
	worksByTicket := make(map[uuid.UUID][]models.TicketWorkView, len(workRows))
	for _, w := range workRows {
		worksByTicket[w.TicketID] = append(worksByTicket[w.TicketID], models.TicketWorkView{
			ComponentID:   w.ComponentID,
			ComponentName: w.ComponentName,
			GroupName:     w.GroupName,
			Action:        w.Action,
			Carrier:       w.Carrier,
			MotorHours:    w.MotorHours,
		})
	}

	result := make([]models.TicketView, 0, len(rows))
	for _, r0 := range rows {
		var downtime *int
		if r0.ClosedAt != nil {
			m := int(r0.ClosedAt.Sub(r0.OpenedAt).Minutes())
			downtime = &m
		}
		result = append(result, models.TicketView{
			ServiceTicket:   r0.ServiceTicket,
			SymptomName:     r0.SymptomName,
			SymptomGroup:    r0.SymptomGroup,
			DowntimeMinutes: downtime,
			Works:           worksByTicket[r0.ID],
		})
	}
	return result, nil
}

// ActiveRecipients возвращает активных получателей уведомлений для типа события.
func (r *PostgresRepository) ActiveRecipients(ctx context.Context, eventType string) ([]models.NotificationRecipient, error) {
	var rows []models.NotificationRecipient
	err := r.db.WithContext(ctx).
		Where("is_active = true AND event_type = ?", eventType).
		Order("created_at").
		Find(&rows).Error
	return rows, err
}

// ListRecipients возвращает всех получателей.
func (r *PostgresRepository) ListRecipients(ctx context.Context) ([]models.NotificationRecipient, error) {
	var rows []models.NotificationRecipient
	err := r.db.WithContext(ctx).Order("created_at").Find(&rows).Error
	return rows, err
}

// CreateRecipient добавляет получателя.
func (r *PostgresRepository) CreateRecipient(ctx context.Context, rec *models.NotificationRecipient) error {
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	if rec.EventType == "" {
		rec.EventType = "service_ticket"
	}
	return r.db.WithContext(ctx).Create(rec).Error
}

// SetRecipientActive включает/выключает получателя.
func (r *PostgresRepository) SetRecipientActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).
		Model(&models.NotificationRecipient{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"is_active": active, "updated_at": time.Now()}).Error
}

// GetCashierUsername возвращает имя (username) кассира по id.
func (r *PostgresRepository) GetCashierUsername(ctx context.Context, cashierID uuid.UUID) (string, error) {
	var username string
	err := r.db.WithContext(ctx).
		Raw(`SELECT username FROM cashiers WHERE id = ?`, cashierID).
		Scan(&username).Error
	return username, err
}

// ComputeBoxMotorHours считает моточасы бокса: (проданные минуты - baseline)/60.
func (r *PostgresRepository) ComputeBoxMotorHours(ctx context.Context, boxNumber int) (int, error) {
	var totalMinutes int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(s.rental_time_minutes + s.extension_time_minutes), 0)
		FROM sessions s JOIN wash_boxes w ON w.id = s.box_id
		WHERE w.number = ? AND s.status = 'complete'`, boxNumber).Scan(&totalMinutes).Error; err != nil {
		return 0, err
	}
	var baseline int64
	// box_maintenance ведётся только для моечных боксов; для остальных baseline = 0.
	r.db.WithContext(ctx).Raw(`SELECT COALESCE(baseline_minutes, 0) FROM box_maintenance WHERE box_number = ?`, boxNumber).Scan(&baseline)

	mh := (totalMinutes - baseline) / 60
	if mh < 0 {
		mh = 0
	}
	return int(mh), nil
}
