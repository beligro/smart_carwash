package service

import (
	washboxRepository "carwash_backend/internal/domain/washbox/repository"
	"carwash_backend/internal/domain/washboxlog/models"
	"carwash_backend/internal/domain/washboxlog/repository"
	"context"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	RecordStatusChange(ctx context.Context, boxID uuid.UUID, oldStatus string, newStatus string, source *string) error
	RecordCoilChange(ctx context.Context, boxID uuid.UUID, action models.ActionType, prevValue *bool, newValue bool, source *string) error
	List(ctx context.Context, boxNumber *int, actorType *string, action *string, since *time.Time, until *time.Time, limit int, offset int) ([]models.WashBoxChangeLog, int64, error)
}

type ServiceImpl struct {
	repo        repository.Repository
	washboxRepo washboxRepository.Repository
}

func NewService(repo repository.Repository, washboxRepo washboxRepository.Repository) *ServiceImpl {
	return &ServiceImpl{repo: repo, washboxRepo: washboxRepo}
}

// deriveActorType пытается извлечь тип актера из контекста
func deriveActorType(ctx context.Context) string {
	// приоритет: dahua/anpr -> роль -> кассир -> уборщик -> пользователь -> system
	if v := ctx.Value("dahua_authenticated"); v != nil {
		if ok, _ := v.(bool); ok {
			return string(models.ActorANPR)
		}
	}
	if v := ctx.Value("role"); v != nil {
		if role, _ := v.(string); role != "" {
			return role
		}
	}
	if v := ctx.Value("cashier_id"); v != nil && v != "" {
		return string(models.ActorCashier)
	}
	if v := ctx.Value("cleaner_id"); v != nil && v != "" {
		return string(models.ActorCleaner)
	}
	if v := ctx.Value("user_id"); v != nil && v != "" {
		return string(models.ActorUser)
	}
	return string(models.ActorSystem)
}

func (s *ServiceImpl) shouldSkip(boxNumber int) bool {
	return boxNumber == 40
}

func (s *ServiceImpl) getBoxNumber(ctx context.Context, boxID uuid.UUID) (int, error) {
	box, err := s.washboxRepo.GetWashBoxByID(ctx, boxID)
	if err != nil || box == nil {
		return 0, err
	}
	return box.Number, nil
}

func (s *ServiceImpl) RecordStatusChange(ctx context.Context, boxID uuid.UUID, oldStatus string, newStatus string, source *string) error {
	// Получаем номер бокса
	boxNumber, err := s.getBoxNumber(ctx, boxID)
	if err != nil {
		return err
	}
	// Пропускаем специальный бокс
	if s.shouldSkip(boxNumber) {
		return nil
	}
	actor := deriveActorType(ctx)
	oldCopy := oldStatus
	newCopy := newStatus
	item := &models.WashBoxChangeLog{
		ID:         uuid.New(),
		BoxID:      boxID,
		BoxNumber:  boxNumber,
		Action:     string(models.ActionStatusChange),
		PrevStatus: &oldCopy,
		NewStatus:  &newCopy,
		ActorType:  actor,
		Source:     source,
		CreatedAt:  time.Now(),
	}
	return s.repo.Insert(ctx, item)
}

func (s *ServiceImpl) RecordCoilChange(ctx context.Context, boxID uuid.UUID, action models.ActionType, prevValue *bool, newValue bool, source *string) error {
	// Получаем номер бокса
	boxNumber, err := s.getBoxNumber(ctx, boxID)
	if err != nil {
		return err
	}
	// Пропускаем специальный бокс
	if s.shouldSkip(boxNumber) {
		return nil
	}
	actor := deriveActorType(ctx)
	newCopy := newValue
	item := &models.WashBoxChangeLog{
		ID:        uuid.New(),
		BoxID:     boxID,
		BoxNumber: boxNumber,
		Action:    string(action),
		PrevValue: prevValue,
		NewValue:  &newCopy,
		ActorType: actor,
		Source:    source,
		CreatedAt: time.Now(),
	}
	return s.repo.Insert(ctx, item)
}

func (s *ServiceImpl) List(ctx context.Context, boxNumber *int, actorType *string, action *string, since *time.Time, until *time.Time, limit int, offset int) ([]models.WashBoxChangeLog, int64, error) {
	return s.repo.List(ctx, boxNumber, actorType, action, since, until, limit, offset)
}


