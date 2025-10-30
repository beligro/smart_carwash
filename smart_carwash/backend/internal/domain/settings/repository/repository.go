package repository

import (
	"carwash_backend/internal/domain/settings/models"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с настройками в базе данных
type Repository interface {
	GetServiceSetting(ctx context.Context, serviceType, settingKey string) (*models.ServiceSetting, error)
	UpdateServiceSetting(ctx context.Context, serviceType, settingKey string, settingValue interface{}) error
	GetAvailableRentalTimes(ctx context.Context, serviceType string) ([]int, error)
	UpdateAvailableRentalTimes(ctx context.Context, serviceType string, times []int) error
}

// RepositoryImpl реализация Repository
type RepositoryImpl struct {
	db *gorm.DB
}

// NewRepository создает новый экземпляр Repository
func NewRepository(db *gorm.DB) *RepositoryImpl {
	return &RepositoryImpl{
		db: db,
	}
}

// GetServiceSetting получает настройку для определенного типа услуги и ключа
func (r *RepositoryImpl) GetServiceSetting(ctx context.Context, serviceType, settingKey string) (*models.ServiceSetting, error) {
	var setting models.ServiceSetting
	var result *gorm.DB
	result = r.db.WithContext(ctx).Where("service_type = ? AND setting_key = ?", serviceType, settingKey).First(&setting)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}

// UpdateServiceSetting обновляет настройку для определенного типа услуги и ключа
func (r *RepositoryImpl) UpdateServiceSetting(ctx context.Context, serviceType, settingKey string, settingValue interface{}) error {
	// Сначала проверяем, существует ли настройка
	setting, err := r.GetServiceSetting(ctx, serviceType, settingKey)
	if err != nil {
		return err
	}

	// Преобразуем значение в JSON
	jsonValue, err := json.Marshal(settingValue)
	if err != nil {
		return err
	}

	if setting == nil {
		// Если настройка не существует, создаем новую
		newSetting := models.ServiceSetting{
			ID:           uuid.New(),
			ServiceType:  serviceType,
			SettingKey:   settingKey,
			SettingValue: jsonValue,
		}
		return r.db.WithContext(ctx).Create(&newSetting).Error
	}

	// Если настройка существует, обновляем ее
	setting.SettingValue = jsonValue
	return r.db.WithContext(ctx).Save(setting).Error
}

// GetAvailableRentalTimes получает доступное время мойки для определенного типа услуги
func (r *RepositoryImpl) GetAvailableRentalTimes(ctx context.Context, serviceType string) ([]int, error) {
	setting, err := r.GetServiceSetting(ctx, serviceType, "available_rental_times")
	if err != nil {
		return nil, err
	}

	if setting == nil {
		// Если настройка не найдена, возвращаем значение по умолчанию
		return []int{5}, nil
	}

	var times []int
	err = json.Unmarshal(setting.SettingValue, &times)
	if err != nil {
		return nil, err
	}

	return times, nil
}

// UpdateAvailableRentalTimes обновляет доступное время мойки для определенного типа услуги
func (r *RepositoryImpl) UpdateAvailableRentalTimes(ctx context.Context, serviceType string, times []int) error {
	return r.UpdateServiceSetting(ctx, serviceType, "available_rental_times", times)
}
