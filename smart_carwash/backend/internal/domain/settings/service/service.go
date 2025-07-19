package service

import (
	"carwash_backend/internal/domain/settings/models"
	"carwash_backend/internal/domain/settings/repository"
	"encoding/json"
)

// Service интерфейс для бизнес-логики настроек
type Service interface {
	GetAvailableRentalTimes(req *models.GetAvailableRentalTimesRequest) (*models.GetAvailableRentalTimesResponse, error)
	UpdateAvailableRentalTimes(req *models.UpdateAvailableRentalTimesRequest) (*models.UpdateAvailableRentalTimesResponse, error)
	GetServicePrice(serviceType, settingKey string) (int, error)
	UpdateServiceSetting(serviceType, settingKey string, value int) error
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo repository.Repository
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository) *ServiceImpl {
	return &ServiceImpl{
		repo: repo,
	}
}

// GetAvailableRentalTimes получает доступное время аренды для определенного типа услуги
func (s *ServiceImpl) GetAvailableRentalTimes(req *models.GetAvailableRentalTimesRequest) (*models.GetAvailableRentalTimesResponse, error) {
	times, err := s.repo.GetAvailableRentalTimes(req.ServiceType)
	if err != nil {
		return nil, err
	}

	return &models.GetAvailableRentalTimesResponse{
		AvailableTimes: times,
	}, nil
}

// UpdateAvailableRentalTimes обновляет доступное время аренды для определенного типа услуги
func (s *ServiceImpl) UpdateAvailableRentalTimes(req *models.UpdateAvailableRentalTimesRequest) (*models.UpdateAvailableRentalTimesResponse, error) {
	err := s.repo.UpdateAvailableRentalTimes(req.ServiceType, req.AvailableTimes)
	if err != nil {
		return nil, err
	}

	return &models.UpdateAvailableRentalTimesResponse{
		Success: true,
	}, nil
}

// GetServicePrice получает цену услуги из настроек
func (s *ServiceImpl) GetServicePrice(serviceType, settingKey string) (int, error) {
	setting, err := s.repo.GetServiceSetting(serviceType, settingKey)
	if err != nil {
		return 0, err
	}

	if setting == nil {
		// Если настройка не найдена, возвращаем значение по умолчанию
		defaultPrices := map[string]map[string]int{
			"wash": {
				"price_per_minute":          1000,
				"chemistry_price_per_minute": 200,
			},
			"air_dry": {
				"price_per_minute":          600,
				"chemistry_price_per_minute": 100,
			},
			"vacuum": {
				"price_per_minute":          400,
				"chemistry_price_per_minute": 50,
			},
		}

		if servicePrices, exists := defaultPrices[serviceType]; exists {
			if price, exists := servicePrices[settingKey]; exists {
				return price, nil
			}
		}

		return 0, nil
	}

	// Парсим JSON значение
	var price int
	err = json.Unmarshal(setting.SettingValue, &price)
	if err != nil {
		return 0, err
	}

	return price, nil
}

// UpdateServiceSetting обновляет настройку для определенного типа услуги и ключа
func (s *ServiceImpl) UpdateServiceSetting(serviceType, settingKey string, value int) error {
	return s.repo.UpdateServiceSetting(serviceType, settingKey, value)
}
