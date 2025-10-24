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

	// Методы для управления доступным временем химии
	GetAvailableChemistryTimes(req *models.GetAvailableChemistryTimesRequest) (*models.GetAvailableChemistryTimesResponse, error)
	AdminGetAvailableChemistryTimes(req *models.AdminGetAvailableChemistryTimesRequest) (*models.AdminGetAvailableChemistryTimesResponse, error)
	AdminUpdateAvailableChemistryTimes(req *models.AdminUpdateAvailableChemistryTimesRequest) (*models.AdminUpdateAvailableChemistryTimesResponse, error)

	// Методы для управления временем уборки
	GetCleaningTimeout() (int, error)
	UpdateCleaningTimeout(timeoutMinutes int) error

	// Методы для управления временем ожидания старта мойки
	GetSessionTimeout() (int, error)
	UpdateSessionTimeout(timeoutMinutes int) error
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

// GetAvailableRentalTimes получает доступное время мойки для определенного типа услуги
func (s *ServiceImpl) GetAvailableRentalTimes(req *models.GetAvailableRentalTimesRequest) (*models.GetAvailableRentalTimesResponse, error) {
	times, err := s.repo.GetAvailableRentalTimes(req.ServiceType)
	if err != nil {
		return nil, err
	}

	return &models.GetAvailableRentalTimesResponse{
		AvailableTimes: times,
	}, nil
}

// UpdateAvailableRentalTimes обновляет доступное время мойки для определенного типа услуги
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

	// Получаем доступное время мойки
	availableTimes, err := s.repo.GetAvailableRentalTimes(req.ServiceType)
	if err != nil {
		return nil, err
	}

	return &models.AdminGetSettingsResponse{
		ServiceType:             req.ServiceType,
		PricePerMinute:          pricePerMinute,
		ChemistryPricePerMinute: chemistryPricePerMinute,
		AvailableRentalTimes:    availableTimes,
	}, nil
}

// GetCleaningTimeout получает время уборки в минутах
func (s *ServiceImpl) GetCleaningTimeout() (int, error) {
	setting, err := s.repo.GetServiceSetting("cleaner", "cleaning_timeout_minutes")
	if err != nil {
		return 3, err // По умолчанию 3 минуты
	}

	if setting == nil {
		return 3, nil // По умолчанию 3 минуты
	}

	var timeout int
	if err := json.Unmarshal(setting.SettingValue, &timeout); err != nil {
		return 3, err // По умолчанию 3 минуты
	}

	return timeout, nil
}

// UpdateCleaningTimeout обновляет время уборки в минутах
func (s *ServiceImpl) UpdateCleaningTimeout(timeoutMinutes int) error {
	return s.repo.UpdateServiceSetting("cleaner", "cleaning_timeout_minutes", timeoutMinutes)
}

// GetSessionTimeout получает время ожидания старта мойки в минутах
func (s *ServiceImpl) GetSessionTimeout() (int, error) {
	setting, err := s.repo.GetServiceSetting("session", "session_timeout_minutes")
	if err != nil {
		return 3, err // По умолчанию 3 минуты
	}

	if setting == nil {
		return 3, nil // По умолчанию 3 минуты
	}

	var timeout int
	if err := json.Unmarshal(setting.SettingValue, &timeout); err != nil {
		return 3, err // По умолчанию 3 минуты
	}

	return timeout, nil
}

// UpdateSessionTimeout обновляет время ожидания старта мойки в минутах
func (s *ServiceImpl) UpdateSessionTimeout(timeoutMinutes int) error {
	return s.repo.UpdateServiceSetting("session", "session_timeout_minutes", timeoutMinutes)
}

// UpdatePrices обновляет цены сервиса (админка)
func (s *ServiceImpl) UpdatePrices(req *models.AdminUpdatePricesRequest) (*models.AdminUpdatePricesResponse, error) {
	// Обновляем цену за минуту
	if err := s.repo.UpdateServiceSetting(req.ServiceType, "price_per_minute", req.PricePerMinute); err != nil {
		return nil, err
	}

	// Обновляем цену химии за минуту
	if req.ChemistryPricePerMinute != nil {
		if err := s.repo.UpdateServiceSetting(req.ServiceType, "chemistry_price_per_minute", *req.ChemistryPricePerMinute); err != nil {
			return nil, err
		}
	}

	return &models.AdminUpdatePricesResponse{
		Success: true,
		Message: "Цены успешно обновлены",
	}, nil
}

// UpdateRentalTimes обновляет доступное время мойки (админка)
func (s *ServiceImpl) UpdateRentalTimes(req *models.AdminUpdateRentalTimesRequest) (*models.AdminUpdateRentalTimesResponse, error) {
	if err := s.repo.UpdateAvailableRentalTimes(req.ServiceType, req.AvailableRentalTimes); err != nil {
		return nil, err
	}

	return &models.AdminUpdateRentalTimesResponse{
		Success: true,
		Message: "Время мойки успешно обновлено",
	}, nil
}

// GetAvailableChemistryTimes получает доступное время химии для определенного типа услуги (публичный метод)
func (s *ServiceImpl) GetAvailableChemistryTimes(req *models.GetAvailableChemistryTimesRequest) (*models.GetAvailableChemistryTimesResponse, error) {
	// Получаем настройку доступного времени химии
	setting, err := s.repo.GetServiceSetting(req.ServiceType, "available_chemistry_times")
	if err != nil {
		return nil, err
	}

	var times []int
	// Если настройка найдена, парсим JSON
	if setting != nil {
		if err := json.Unmarshal(setting.SettingValue, &times); err != nil {
			return nil, err
		}
	} else {
		// Значения по умолчанию
		times = []int{3, 4, 5}
	}

	return &models.GetAvailableChemistryTimesResponse{
		AvailableChemistryTimes: times,
	}, nil
}

// AdminGetAvailableChemistryTimes получает доступное время химии для определенного типа услуги (админка)
func (s *ServiceImpl) AdminGetAvailableChemistryTimes(req *models.AdminGetAvailableChemistryTimesRequest) (*models.AdminGetAvailableChemistryTimesResponse, error) {
	// Получаем настройку доступного времени химии
	setting, err := s.repo.GetServiceSetting(req.ServiceType, "available_chemistry_times")
	if err != nil {
		return nil, err
	}

	var times []int
	// Если настройка найдена, парсим JSON
	if setting != nil {
		if err := json.Unmarshal(setting.SettingValue, &times); err != nil {
			return nil, err
		}
	} else {
		// Значения по умолчанию
		times = []int{3, 4, 5}
	}

	return &models.AdminGetAvailableChemistryTimesResponse{
		ServiceType:             req.ServiceType,
		AvailableChemistryTimes: times,
	}, nil
}

// AdminUpdateAvailableChemistryTimes обновляет доступное время химии для определенного типа услуги (админка)
func (s *ServiceImpl) AdminUpdateAvailableChemistryTimes(req *models.AdminUpdateAvailableChemistryTimesRequest) (*models.AdminUpdateAvailableChemistryTimesResponse, error) {
	// Обновляем настройку (UpdateServiceSetting сам сделает Marshal)
	if err := s.repo.UpdateServiceSetting(req.ServiceType, "available_chemistry_times", req.AvailableChemistryTimes); err != nil {
		return nil, err
	}

	return &models.AdminUpdateAvailableChemistryTimesResponse{
		Success: true,
		Message: "Доступное время химии успешно обновлено",
	}, nil
}
