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
	GetSettings(req *models.AdminGetSettingsRequest) (*models.AdminGetSettingsResponse, error)
	UpdatePrices(req *models.AdminUpdatePricesRequest) (*models.AdminUpdatePricesResponse, error)
	UpdateRentalTimes(req *models.AdminUpdateRentalTimesRequest) (*models.AdminUpdateRentalTimesResponse, error)
	
	// Методы для управления временем доступности химии
	GetChemistryTimeout(req *models.AdminGetChemistryTimeoutRequest) (*models.AdminGetChemistryTimeoutResponse, error)
	UpdateChemistryTimeout(req *models.AdminUpdateChemistryTimeoutRequest) (*models.AdminUpdateChemistryTimeoutResponse, error)
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

// GetSettings получает все настройки сервиса (админка)
func (s *ServiceImpl) GetSettings(req *models.AdminGetSettingsRequest) (*models.AdminGetSettingsResponse, error) {
	// Получаем цену за минуту
	pricePerMinuteSetting, err := s.repo.GetServiceSetting(req.ServiceType, "price_per_minute")
	if err != nil {
		return nil, err
	}

	var pricePerMinute int
	if pricePerMinuteSetting != nil {
		if err := json.Unmarshal(pricePerMinuteSetting.SettingValue, &pricePerMinute); err != nil {
			return nil, err
		}
	}

	// Получаем цену химии за минуту
	chemistryPricePerMinuteSetting, err := s.repo.GetServiceSetting(req.ServiceType, "chemistry_price_per_minute")
	if err != nil {
		return nil, err
	}

	var chemistryPricePerMinute int
	if chemistryPricePerMinuteSetting != nil {
		if err := json.Unmarshal(chemistryPricePerMinuteSetting.SettingValue, &chemistryPricePerMinute); err != nil {
			return nil, err
		}
	}

	// Получаем доступное время аренды
	availableTimes, err := s.repo.GetAvailableRentalTimes(req.ServiceType)
	if err != nil {
		return nil, err
	}

	return &models.AdminGetSettingsResponse{
		ServiceType:           req.ServiceType,
		PricePerMinute:       pricePerMinute,
		ChemistryPricePerMinute: chemistryPricePerMinute,
		AvailableRentalTimes: availableTimes,
	}, nil
}

// UpdatePrices обновляет цены сервиса (админка)
func (s *ServiceImpl) UpdatePrices(req *models.AdminUpdatePricesRequest) (*models.AdminUpdatePricesResponse, error) {
	// Обновляем цену за минуту
	if err := s.repo.UpdateServiceSetting(req.ServiceType, "price_per_minute", req.PricePerMinute); err != nil {
		return nil, err
	}

	// Обновляем цену химии за минуту
	if err := s.repo.UpdateServiceSetting(req.ServiceType, "chemistry_price_per_minute", req.ChemistryPricePerMinute); err != nil {
		return nil, err
	}

	return &models.AdminUpdatePricesResponse{
		Success: true,
		Message: "Цены успешно обновлены",
	}, nil
}

// UpdateRentalTimes обновляет доступное время аренды (админка)
func (s *ServiceImpl) UpdateRentalTimes(req *models.AdminUpdateRentalTimesRequest) (*models.AdminUpdateRentalTimesResponse, error) {
	if err := s.repo.UpdateAvailableRentalTimes(req.ServiceType, req.AvailableRentalTimes); err != nil {
		return nil, err
	}

	return &models.AdminUpdateRentalTimesResponse{
		Success: true,
		Message: "Время аренды успешно обновлено",
	}, nil
}

// GetChemistryTimeout получает время доступности кнопки химии
func (s *ServiceImpl) GetChemistryTimeout(req *models.AdminGetChemistryTimeoutRequest) (*models.AdminGetChemistryTimeoutResponse, error) {
	setting, err := s.repo.GetServiceSetting(req.ServiceType, "chemistry_enable_timeout_minutes")
	if err != nil {
		return nil, err
	}

	var timeoutMinutes int
	if setting != nil {
		if err := json.Unmarshal(setting.SettingValue, &timeoutMinutes); err != nil {
			return nil, err
		}
	} else {
		// Значение по умолчанию
		timeoutMinutes = 10
	}

	return &models.AdminGetChemistryTimeoutResponse{
		ServiceType:              req.ServiceType,
		ChemistryEnableTimeoutMinutes: timeoutMinutes,
	}, nil
}

// UpdateChemistryTimeout обновляет время доступности кнопки химии
func (s *ServiceImpl) UpdateChemistryTimeout(req *models.AdminUpdateChemistryTimeoutRequest) (*models.AdminUpdateChemistryTimeoutResponse, error) {
	if err := s.repo.UpdateServiceSetting(req.ServiceType, "chemistry_enable_timeout_minutes", req.ChemistryEnableTimeoutMinutes); err != nil {
		return nil, err
	}

	return &models.AdminUpdateChemistryTimeoutResponse{
		Success: true,
		Message: "Время доступности кнопки химии успешно обновлено",
	}, nil
}
