package service

import (
	"carwash_backend/internal/domain/settings/models"
	"carwash_backend/internal/domain/settings/repository"
	"context"
	"encoding/json"
)

// Service интерфейс для бизнес-логики настроек
type Service interface {
	GetAvailableRentalTimes(ctx context.Context, req *models.GetAvailableRentalTimesRequest) (*models.GetAvailableRentalTimesResponse, error)
	UpdateAvailableRentalTimes(ctx context.Context, req *models.UpdateAvailableRentalTimesRequest) (*models.UpdateAvailableRentalTimesResponse, error)
	GetSettings(ctx context.Context, req *models.AdminGetSettingsRequest) (*models.AdminGetSettingsResponse, error)
	UpdatePrices(ctx context.Context, req *models.AdminUpdatePricesRequest) (*models.AdminUpdatePricesResponse, error)
	UpdateRentalTimes(ctx context.Context, req *models.AdminUpdateRentalTimesRequest) (*models.AdminUpdateRentalTimesResponse, error)

	// Методы для управления доступным временем химии
	GetAvailableChemistryTimes(ctx context.Context, req *models.GetAvailableChemistryTimesRequest) (*models.GetAvailableChemistryTimesResponse, error)
	AdminGetAvailableChemistryTimes(ctx context.Context, req *models.AdminGetAvailableChemistryTimesRequest) (*models.AdminGetAvailableChemistryTimesResponse, error)
	AdminUpdateAvailableChemistryTimes(ctx context.Context, req *models.AdminUpdateAvailableChemistryTimesRequest) (*models.AdminUpdateAvailableChemistryTimesResponse, error)

	// Методы для управления временем уборки
	GetCleaningTimeout(ctx context.Context) (int, error)
	UpdateCleaningTimeout(ctx context.Context, timeoutMinutes int) error

	// Методы для управления временем ожидания старта мойки
	GetSessionTimeout(ctx context.Context) (int, error)
	UpdateSessionTimeout(ctx context.Context, timeoutMinutes int) error

	// Методы для управления временем блокировки бокса после завершения сессии
	GetCooldownTimeout(ctx context.Context) (int, error)
	UpdateCooldownTimeout(ctx context.Context, timeoutMinutes int) error
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
func (s *ServiceImpl) GetAvailableRentalTimes(ctx context.Context, req *models.GetAvailableRentalTimesRequest) (*models.GetAvailableRentalTimesResponse, error) {
	times, err := s.repo.GetAvailableRentalTimes(ctx, req.ServiceType)
	if err != nil {
		return nil, err
	}

	return &models.GetAvailableRentalTimesResponse{
		AvailableTimes: times,
	}, nil
}

// UpdateAvailableRentalTimes обновляет доступное время мойки для определенного типа услуги
func (s *ServiceImpl) UpdateAvailableRentalTimes(ctx context.Context, req *models.UpdateAvailableRentalTimesRequest) (*models.UpdateAvailableRentalTimesResponse, error) {
	err := s.repo.UpdateAvailableRentalTimes(ctx, req.ServiceType, req.AvailableTimes)
	if err != nil {
		return nil, err
	}

	return &models.UpdateAvailableRentalTimesResponse{
		Success: true,
	}, nil
}

// GetSettings получает все настройки сервиса (админка)
func (s *ServiceImpl) GetSettings(ctx context.Context, req *models.AdminGetSettingsRequest) (*models.AdminGetSettingsResponse, error) {
	// Получаем цену за минуту
	pricePerMinuteSetting, err := s.repo.GetServiceSetting(ctx, req.ServiceType, "price_per_minute")
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
	chemistryPricePerMinuteSetting, err := s.repo.GetServiceSetting(ctx, req.ServiceType, "chemistry_price_per_minute")
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
	availableTimes, err := s.repo.GetAvailableRentalTimes(ctx, req.ServiceType)
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
func (s *ServiceImpl) GetCleaningTimeout(ctx context.Context) (int, error) {
	setting, err := s.repo.GetServiceSetting(ctx, "cleaner", "cleaning_timeout_minutes")
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
func (s *ServiceImpl) UpdateCleaningTimeout(ctx context.Context, timeoutMinutes int) error {
	return s.repo.UpdateServiceSetting(ctx, "cleaner", "cleaning_timeout_minutes", timeoutMinutes)
}

// GetSessionTimeout получает время ожидания старта мойки в минутах
func (s *ServiceImpl) GetSessionTimeout(ctx context.Context) (int, error) {
	setting, err := s.repo.GetServiceSetting(ctx, "session", "session_timeout_minutes")
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
func (s *ServiceImpl) UpdateSessionTimeout(ctx context.Context, timeoutMinutes int) error {
	return s.repo.UpdateServiceSetting(ctx, "session", "session_timeout_minutes", timeoutMinutes)
}

// GetCooldownTimeout получает время блокировки бокса после завершения сессии в минутах
func (s *ServiceImpl) GetCooldownTimeout(ctx context.Context) (int, error) {
	setting, err := s.repo.GetServiceSetting(ctx, "session", "cooldown_timeout_minutes")
	if err != nil {
		return 5, err // По умолчанию 5 минут
	}

	if setting == nil {
		return 5, nil // По умолчанию 5 минут
	}

	var timeout int
	if err := json.Unmarshal(setting.SettingValue, &timeout); err != nil {
		return 5, err // По умолчанию 5 минут
	}

	return timeout, nil
}

// UpdateCooldownTimeout обновляет время блокировки бокса после завершения сессии в минутах
func (s *ServiceImpl) UpdateCooldownTimeout(ctx context.Context, timeoutMinutes int) error {
	return s.repo.UpdateServiceSetting(ctx, "session", "cooldown_timeout_minutes", timeoutMinutes)
}

// UpdatePrices обновляет цены сервиса (админка)
func (s *ServiceImpl) UpdatePrices(ctx context.Context, req *models.AdminUpdatePricesRequest) (*models.AdminUpdatePricesResponse, error) {
	// Обновляем цену за минуту
	if err := s.repo.UpdateServiceSetting(ctx, req.ServiceType, "price_per_minute", req.PricePerMinute); err != nil {
		return nil, err
	}

	// Обновляем цену химии за минуту
	if req.ChemistryPricePerMinute != nil {
		if err := s.repo.UpdateServiceSetting(ctx, req.ServiceType, "chemistry_price_per_minute", *req.ChemistryPricePerMinute); err != nil {
			return nil, err
		}
	}

	return &models.AdminUpdatePricesResponse{
		Success: true,
		Message: "Цены успешно обновлены",
	}, nil
}

// UpdateRentalTimes обновляет доступное время мойки (админка)
func (s *ServiceImpl) UpdateRentalTimes(ctx context.Context, req *models.AdminUpdateRentalTimesRequest) (*models.AdminUpdateRentalTimesResponse, error) {
	if err := s.repo.UpdateAvailableRentalTimes(ctx, req.ServiceType, req.AvailableRentalTimes); err != nil {
		return nil, err
	}

	return &models.AdminUpdateRentalTimesResponse{
		Success: true,
		Message: "Время мойки успешно обновлено",
	}, nil
}

// GetAvailableChemistryTimes получает доступное время химии для определенного типа услуги (публичный метод)
func (s *ServiceImpl) GetAvailableChemistryTimes(ctx context.Context, req *models.GetAvailableChemistryTimesRequest) (*models.GetAvailableChemistryTimesResponse, error) {
	// Получаем настройку доступного времени химии
	setting, err := s.repo.GetServiceSetting(ctx, req.ServiceType, "available_chemistry_times")
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
func (s *ServiceImpl) AdminGetAvailableChemistryTimes(ctx context.Context, req *models.AdminGetAvailableChemistryTimesRequest) (*models.AdminGetAvailableChemistryTimesResponse, error) {
	// Получаем настройку доступного времени химии
	setting, err := s.repo.GetServiceSetting(ctx, req.ServiceType, "available_chemistry_times")
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
func (s *ServiceImpl) AdminUpdateAvailableChemistryTimes(ctx context.Context, req *models.AdminUpdateAvailableChemistryTimesRequest) (*models.AdminUpdateAvailableChemistryTimesResponse, error) {
	// Обновляем настройку (UpdateServiceSetting сам сделает Marshal)
	if err := s.repo.UpdateServiceSetting(ctx, req.ServiceType, "available_chemistry_times", req.AvailableChemistryTimes); err != nil {
		return nil, err
	}

	return &models.AdminUpdateAvailableChemistryTimesResponse{
		Success: true,
		Message: "Доступное время химии успешно обновлено",
	}, nil
}
