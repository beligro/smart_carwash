package repository

import (
	"carwash_backend/internal/domain/settings/models"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с настройками в базе данных
type Repository interface {
	GetServiceSetting(serviceType, settingKey string) (*models.ServiceSetting, error)
	UpdateServiceSetting(serviceType, settingKey string, settingValue interface{}) error
	GetAvailableRentalTimes(serviceType string) ([]int, error)
	UpdateAvailableRentalTimes(serviceType string, times []int) error
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
func (r *RepositoryImpl) GetServiceSetting(serviceType, settingKey string) (*models.ServiceSetting, error) {
	var setting models.ServiceSetting
	result := r.db.Where("service_type = ? AND setting_key = ?", serviceType, settingKey).First(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &setting, nil
}

// UpdateServiceSetting обновляет настройку для определенного типа услуги и ключа
func (r *RepositoryImpl) UpdateServiceSetting(serviceType, settingKey string, settingValue interface{}) error {
	// Сначала проверяем, существует ли настройка
	setting, err := r.GetServiceSetting(serviceType, settingKey)
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
		result := r.db.Create(&newSetting)
		return result.Error
	}

	// Если настройка существует, обновляем ее
	setting.SettingValue = jsonValue
	result := r.db.Save(setting)
	return result.Error
}

// GetAvailableRentalTimes получает доступное время мойки для определенного типа услуги
func (r *RepositoryImpl) GetAvailableRentalTimes(serviceType string) ([]int, error) {
	setting, err := r.GetServiceSetting(serviceType, "available_rental_times")
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
func (r *RepositoryImpl) UpdateAvailableRentalTimes(serviceType string, times []int) error {
	return r.UpdateServiceSetting(serviceType, "available_rental_times", times)
}
