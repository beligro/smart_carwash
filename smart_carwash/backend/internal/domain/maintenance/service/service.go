package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"carwash_backend/internal/domain/maintenance/models"
	"carwash_backend/internal/domain/maintenance/repository"
	washboxModels "carwash_backend/internal/domain/washbox/models"
	"carwash_backend/internal/logger"

	"github.com/google/uuid"
)

// Notifier отправляет push-уведомления (реализуется telegram-ботом).
type Notifier interface {
	SendMaintenanceAlert(chatID int64, text string) error
}

// WashboxService — нужный маинтенансу срез сервиса боксов.
type WashboxService interface {
	GetWashBoxByID(ctx context.Context, id uuid.UUID) (*washboxModels.WashBox, error)
	CashierSetMaintenance(ctx context.Context, req *washboxModels.CashierSetMaintenanceRequest) (*washboxModels.CashierSetMaintenanceResponse, error)
	UpdateWashBoxStatus(ctx context.Context, id uuid.UUID, status string) error
}

// Service бизнес-логика сервисных нарядов.
type Service interface {
	GetSymptoms(ctx context.Context, boxNumber int) ([]models.SymptomType, error)
	GetCatalog(ctx context.Context, boxNumber int) ([]models.CatalogGroup, error)
	OpenTicket(ctx context.Context, cashierID *uuid.UUID, req *models.OpenTicketRequest) (*models.ServiceTicket, error)
	CloseTicket(ctx context.Context, ticketID uuid.UUID, closedBy string, req *models.CloseTicketRequest) (*models.ServiceTicket, error)
	ListTickets(ctx context.Context, status *string, boxNumber *int, since *time.Time, limit int) ([]models.TicketView, error)
	ReconcileOrphanTickets(ctx context.Context) error

	ListRecipients(ctx context.Context) ([]models.NotificationRecipient, error)
	CreateRecipient(ctx context.Context, req *models.CreateRecipientRequest) (*models.NotificationRecipient, error)
	SetRecipientActive(ctx context.Context, id uuid.UUID, active bool) error
}

// ServiceImpl реализация Service.
type ServiceImpl struct {
	repo       repository.Repository
	washboxSvc WashboxService
	notifier   Notifier
}

// NewService создаёт сервис. notifier может быть nil (push отключён).
func NewService(repo repository.Repository, washboxSvc WashboxService, notifier Notifier) *ServiceImpl {
	return &ServiceImpl{repo: repo, washboxSvc: washboxSvc, notifier: notifier}
}

// BoxType определяет тип бокса по номеру: мойка 1-8,11-18; пылесос 9,10,19,20; воздух 21,22,23.
func BoxType(number int) string {
	switch number {
	case 9, 10, 19, 20:
		return models.BoxTypeVacuum
	case 21, 22, 23:
		return models.BoxTypeAir
	default:
		return models.BoxTypeWash
	}
}

// GetSymptoms возвращает симптомы для типа бокса по его номеру.
func (s *ServiceImpl) GetSymptoms(ctx context.Context, boxNumber int) ([]models.SymptomType, error) {
	return s.repo.GetSymptomsByBoxType(ctx, BoxType(boxNumber))
}

// GetCatalog возвращает справочник работ для типа бокса по его номеру.
func (s *ServiceImpl) GetCatalog(ctx context.Context, boxNumber int) ([]models.CatalogGroup, error) {
	return s.repo.GetCatalog(ctx, BoxType(boxNumber))
}

// OpenTicket ставит бокс в сервис и открывает наряд, затем шлёт push получателям.
func (s *ServiceImpl) OpenTicket(ctx context.Context, cashierID *uuid.UUID, req *models.OpenTicketRequest) (*models.ServiceTicket, error) {
	// Симптом обязателен: нельзя ставить бокс в сервис без указания причины.
	// Проверяем ДО перевода бокса в maintenance, чтобы не оставить бокс без наряда.
	if req.SymptomID == nil {
		return nil, errors.New("укажите причину (симптом) для постановки в сервис")
	}
	symptom, err := s.repo.GetSymptomByID(ctx, *req.SymptomID)
	if err != nil || symptom == nil {
		return nil, errors.New("указан несуществующий симптом")
	}

	box, err := s.washboxSvc.GetWashBoxByID(ctx, req.BoxID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	// Переводим бокс в сервис (валидация "только свободные" внутри).
	if _, err := s.washboxSvc.CashierSetMaintenance(ctx, &washboxModels.CashierSetMaintenanceRequest{ID: req.BoxID}); err != nil {
		return nil, err
	}

	openedBy := ""
	if cashierID != nil {
		if name, e := s.repo.GetCashierUsername(ctx, *cashierID); e == nil {
			openedBy = name
		}
	}

	boxID := box.ID
	ticket := &models.ServiceTicket{
		BoxID:             &boxID,
		BoxNumber:         box.Number,
		BoxType:           BoxType(box.Number),
		Status:            models.TicketStatusOpen,
		IsBreakdown:       symptom.IsBreakdown,
		OpenedAt:          time.Now(),
		OpenedBy:          openedBy,
		OpenedByCashierID: cashierID,
		SymptomID:         req.SymptomID,
		CashierComment:    req.Comment,
	}
	if err := s.repo.CreateTicket(ctx, ticket); err != nil {
		// Бокс уже в сервисе; наряд не записан — логируем, но не откатываем статус.
		logger.Printf("Ошибка создания наряда для бокса %d: %v", box.Number, err)
		return nil, fmt.Errorf("бокс переведён в сервис, но наряд не сохранён: %v", err)
	}

	go s.pushTicketOpened(ticket)

	return ticket, nil
}

// pushTicketOpened рассылает уведомление активным получателям (в фоне).
func (s *ServiceImpl) pushTicketOpened(ticket *models.ServiceTicket) {
	if s.notifier == nil {
		return
	}
	ctx := context.Background()

	symptom := "не указан"
	if ticket.SymptomID != nil {
		if syms, err := s.repo.GetSymptomsByBoxType(ctx, ticket.BoxType); err == nil {
			for _, sy := range syms {
				if sy.ID == *ticket.SymptomID {
					symptom = sy.Name
					break
				}
			}
		}
	}

	cashier := ticket.OpenedBy
	if cashier == "" {
		cashier = "—"
	}
	text := fmt.Sprintf("🔧 <b>Бокс №%d → сервис</b>\nКассир: %s\nСимптом: %s", ticket.BoxNumber, cashier, symptom)
	if ticket.CashierComment != "" {
		text += "\nКоммент: " + ticket.CashierComment
	}

	recipients, err := s.repo.ActiveRecipients(ctx, "service_ticket")
	if err != nil {
		logger.Printf("Ошибка получения получателей уведомлений: %v", err)
		return
	}
	for _, r := range recipients {
		if err := s.notifier.SendMaintenanceAlert(r.ChatID, text); err != nil {
			logger.Printf("Ошибка отправки push получателю %s (%d): %v", r.Name, r.ChatID, err)
		}
	}
}

// CloseTicket закрывает наряд: пишет работы с моточасами и возвращает бокс в работу.
func (s *ServiceImpl) CloseTicket(ctx context.Context, ticketID uuid.UUID, closedBy string, req *models.CloseTicketRequest) (*models.ServiceTicket, error) {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, errors.New("наряд не найден")
	}
	if ticket.Status != models.TicketStatusOpen {
		return nil, errors.New("наряд уже закрыт")
	}
	// Для поломок закрывать наряд можно только с указанием выполненных работ (для статистики).
	// Не-поломки (напр. «Бокс заблокирован») закрываются без работ и в статистику не идут.
	if ticket.IsBreakdown && len(req.Works) == 0 {
		return nil, errors.New("укажите хотя бы одну выполненную работу")
	}

	// Моточасы бокса на момент закрытия (Фаза 1: одометр бокса для всех носителей;
	// TODO Фаза 2 — отдельный одометр аппарата для узлов помпы/мотора).
	boxMotorHours, _ := s.repo.ComputeBoxMotorHours(ctx, ticket.BoxNumber)

	works := make([]models.TicketWork, 0, len(req.Works))
	for _, w := range req.Works {
		carrier, _ := s.repo.GetComponentCarrier(ctx, w.ComponentID)
		mh := boxMotorHours
		works = append(works, models.TicketWork{
			TicketID:    ticket.ID,
			ComponentID: w.ComponentID,
			Action:      w.Action,
			Carrier:     carrier,
			MotorHours:  &mh,
		})
	}

	now := time.Now()
	ticket.ClosedAt = &now
	ticket.ClosedBy = closedBy
	ticket.MasterComment = req.MasterComment
	ticket.Status = models.TicketStatusClosed

	if err := s.repo.CloseTicket(ctx, ticket, works); err != nil {
		return nil, err
	}

	// Возвращаем бокс в работу.
	if ticket.BoxID != nil {
		if err := s.washboxSvc.UpdateWashBoxStatus(ctx, *ticket.BoxID, washboxModels.StatusFree); err != nil {
			logger.Printf("Наряд %s закрыт, но не удалось вернуть бокс %d в работу: %v", ticket.ID, ticket.BoxNumber, err)
		}
	}

	return ticket, nil
}

// ReconcileOrphanTickets закрывает открытые наряды, чей бокс уже не в сервисе
// (вышел из maintenance вне раздела нарядов) — страховка от рассинхрона.
func (s *ServiceImpl) ReconcileOrphanTickets(ctx context.Context) error {
	tickets, err := s.repo.GetAllOpenTickets(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	for i := range tickets {
		t := tickets[i]
		if t.BoxID == nil {
			continue
		}
		box, err := s.washboxSvc.GetWashBoxByID(ctx, *t.BoxID)
		if err != nil {
			continue
		}
		if box.Status == washboxModels.StatusMaintenance {
			continue // бокс всё ещё в сервисе — наряд валиден
		}
		t.ClosedAt = &now
		t.ClosedBy = "Авто (бокс вернулся в работу)"
		t.MasterComment = "Наряд закрыт автоматически: бокс вышел из сервиса вне раздела нарядов"
		t.Status = models.TicketStatusClosed
		if err := s.repo.CloseTicket(ctx, &t, nil); err != nil {
			logger.Printf("Reconcile: не удалось авто-закрыть наряд %s: %v", t.ID, err)
		}
	}
	return nil
}

// ListTickets возвращает журнал нарядов.
func (s *ServiceImpl) ListTickets(ctx context.Context, status *string, boxNumber *int, since *time.Time, limit int) ([]models.TicketView, error) {
	if limit <= 0 {
		limit = 200
	}
	return s.repo.ListTickets(ctx, status, boxNumber, since, limit)
}

// ListRecipients возвращает всех получателей уведомлений.
func (s *ServiceImpl) ListRecipients(ctx context.Context) ([]models.NotificationRecipient, error) {
	return s.repo.ListRecipients(ctx)
}

// CreateRecipient добавляет получателя уведомлений.
func (s *ServiceImpl) CreateRecipient(ctx context.Context, req *models.CreateRecipientRequest) (*models.NotificationRecipient, error) {
	rec := &models.NotificationRecipient{
		Name:      req.Name,
		ChatID:    req.ChatID,
		EventType: "service_ticket",
		IsActive:  true,
	}
	if err := s.repo.CreateRecipient(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// SetRecipientActive включает/выключает получателя.
func (s *ServiceImpl) SetRecipientActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.repo.SetRecipientActive(ctx, id, active)
}
