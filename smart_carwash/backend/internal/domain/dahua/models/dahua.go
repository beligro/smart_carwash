package models

import (
	"encoding/xml"
	"time"
)

// DahuaWebhookRequest представляет входящий webhook от камеры Dahua в формате ITSAPI XML
type DahuaWebhookRequest struct {
	XMLName          xml.Name `xml:"EventNotificationAlert"`
	IPAddress        string   `xml:"ipAddress"`           // IP адрес камеры
	PortNo           int      `xml:"portNo"`              // Порт камеры
	Protocol         string   `xml:"protocol"`            // Протокол (HTTP)
	MacAddress       string   `xml:"macAddress"`          // MAC адрес камеры
	ChannelID        int      `xml:"channelID"`           // ID канала
	DateTime         string   `xml:"dateTime"`            // Время события
	ActivePostCount  int      `xml:"activePostCount"`     // Количество активных постов
	EventType        string   `xml:"eventType"`           // Тип события (ANPR)
	EventState       string   `xml:"eventState"`          // Состояние события (active/inactive)
	EventDescription string   `xml:"eventDescription"`    // Описание события
	LicensePlate     string   `xml:"licensePlate"`        // Номер автомобиля
	Confidence       int      `xml:"confidence"`          // Уровень уверенности распознавания
	Direction        string   `xml:"direction"`           // Направление движения (in/out)
	ImagePath        string   `xml:"imagePath,omitempty"` // Путь к изображению (опционально)
}

// DahuaWebhookRequestJSON представляет входящий webhook от камеры Dahua в JSON формате
type DahuaWebhookRequestJSON struct {
	Picture struct {
		Plate struct {
			BoundingBox []int  `json:"BoundingBox"` // Координаты номера
			Channel     int    `json:"Channel"`     // Канал
			IsExist     bool   `json:"IsExist"`     // Существует ли номер
			PlateColor  string `json:"PlateColor"`  // Цвет номера
			PlateNumber string `json:"PlateNumber"` // Номер автомобиля
			PlateType   string `json:"PlateType"`   // Тип номера
			Region      string `json:"Region"`      // Регион
			UploadNum   int    `json:"UploadNum"`   // Номер загрузки
		} `json:"Plate"`
		SnapInfo struct {
			AllowUser        bool   `json:"AllowUser"`        // Разрешено пользователю
			AllowUserEndTime string `json:"AllowUserEndTime"` // Время окончания разрешения
			DefenceCode      string `json:"DefenceCode"`      // Код защиты
			DeviceID         string `json:"DeviceID"`         // ID устройства
			InCarPeopleNum   int    `json:"InCarPeopleNum"`   // Количество людей в машине
			LanNo            int    `json:"LanNo"`            // Номер полосы
			OpenStrobe       bool   `json:"OpenStrobe"`       // Открыта ли вспышка
		} `json:"SnapInfo"`
		Vehicle struct {
			VehicleBoundingBox []int  `json:"VehicleBoundingBox"` // Координаты автомобиля
			VehicleSeries      string `json:"VehicleSeries"`      // Серия автомобиля
		} `json:"Vehicle"`
	} `json:"Picture"`
}

// DahuaWebhookResponse представляет ответ на webhook от Dahua в формате ITSAPI XML
type DahuaWebhookResponse struct {
	XMLName xml.Name `xml:"Response"`
	Result  string   `xml:"result"`  // OK или ERROR
	Message string   `xml:"message"` // Сообщение о результате
}

// DahuaWebhookResponseJSON представляет ответ на webhook от Dahua в JSON формате (для обратной совместимости)
type DahuaWebhookResponseJSON struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ProcessANPREventRequest представляет запрос на обработку ANPR события
type ProcessANPREventRequest struct {
	LicensePlate string `json:"license_plate" binding:"required"`
	Direction    string `json:"direction" binding:"required"`
	Confidence   int    `json:"confidence"`
	EventType    string `json:"event_type"`
	CaptureTime  string `json:"capture_time"`
	ImagePath    string `json:"image_path"`
}

// ProcessANPREventResponse представляет ответ на обработку ANPR события
type ProcessANPREventResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	UserFound     bool   `json:"user_found"`
	SessionFound  bool   `json:"session_found"`
	SessionID     string `json:"session_id,omitempty"`
	SessionStatus string `json:"session_status,omitempty"`
}

// ValidateDirection проверяет корректность направления движения
func (req *DahuaWebhookRequest) ValidateDirection() bool {
	return req.Direction == "in" || req.Direction == "out"
}

// ValidateEventType проверяет корректность типа события
func (req *DahuaWebhookRequest) ValidateEventType() bool {
	return req.EventType == "ANPR"
}

// ValidateEventState проверяет корректность состояния события
func (req *DahuaWebhookRequest) ValidateEventState() bool {
	return req.EventState == "active" || req.EventState == "inactive"
}

// ToProcessRequest преобразует XML запрос в внутренний формат для обработки
func (req *DahuaWebhookRequest) ToProcessRequest() *ProcessANPREventRequest {
	return &ProcessANPREventRequest{
		LicensePlate: req.LicensePlate,
		Direction:    req.Direction,
		Confidence:   req.Confidence,
		EventType:    req.EventType,
		CaptureTime:  req.DateTime,
		ImagePath:    req.ImagePath,
	}
}

// ValidatePlateNumber проверяет, что номер автомобиля существует
func (req *DahuaWebhookRequestJSON) ValidatePlateNumber() bool {
	return req.Picture.Plate.PlateNumber != ""
}

// GetPlateNumber возвращает номер автомобиля
func (req *DahuaWebhookRequestJSON) GetPlateNumber() string {
	return req.Picture.Plate.PlateNumber
}

// GetDeviceID возвращает ID устройства
func (req *DahuaWebhookRequestJSON) GetDeviceID() string {
	return req.Picture.SnapInfo.DeviceID
}

// ToProcessRequest преобразует JSON запрос в внутренний формат для обработки
func (req *DahuaWebhookRequestJSON) ToProcessRequest() *ProcessANPREventRequest {
	return &ProcessANPREventRequest{
		LicensePlate: req.Picture.Plate.PlateNumber,
		Direction:    "out", // Любой запрос означает выезд
		Confidence:   100,   // Считаем что камера уверена
		EventType:    "ANPR",
		CaptureTime:  time.Now().Format("2006-01-02T15:04:05"),
		ImagePath:    "", // Нет пути к изображению в новой структуре
	}
}

// DahuaDeviceRegistration представляет данные регистрации устройства от камеры Dahua
type DahuaDeviceRegistration struct {
	DeviceID     string `json:"DeviceID"`     // ID устройства
	DeviceModel  string `json:"DeviceModel"`  // Модель устройства
	DeviceName   string `json:"DeviceName"`   // Имя устройства
	DeviceType   string `json:"DeviceType"`   // Тип устройства
	IPAddress    string `json:"IPAddress"`    // IP адрес
	IPv6Address  string `json:"IPv6Address"`  // IPv6 адрес
	MACAddress   string `json:"MACAddress"`   // MAC адрес
	Manufacturer string `json:"Manufacturer"` // Производитель
}

// DeviceRegistrationResponse представляет ответ на регистрацию устройства
type DeviceRegistrationResponse struct {
	Result    string `json:"Result"`    // OK или Error
	DeviceID  string `json:"DeviceID"`  // ID устройства
	Message   string `json:"Message"`   // Сообщение
	Timestamp string `json:"Timestamp"` // Время ответа
	ServerID  string `json:"ServerID"`  // ID сервера
	Status    string `json:"Status"`    // Статус
}

// IsExitEvent проверяет, является ли событие выездом
func (r *DahuaWebhookRequest) IsExitEvent() bool {
	return r.Direction == "out"
}

// ParseCaptureTime парсит время захвата из строки
func (r *DahuaWebhookRequest) ParseCaptureTime() (time.Time, error) {
	// Формат времени от Dahua: "2025-10-20T14:05:33"
	return time.Parse("2006-01-02T15:04:05", r.DateTime)
}
