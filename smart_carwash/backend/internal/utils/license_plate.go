package utils

import (
	"regexp"
	"strings"
)

// EnglishToRussian маппинг английских букв на русские для госномеров
var EnglishToRussian = map[rune]rune{
	'A': 'А', 'B': 'В', 'E': 'Е', 'K': 'К', 'M': 'М',
	'H': 'Н', 'O': 'О', 'P': 'Р', 'C': 'С', 'T': 'Т',
	'Y': 'У', 'X': 'Х',
	'a': 'А', 'b': 'В', 'e': 'Е', 'k': 'К', 'm': 'М',
	'h': 'Н', 'o': 'О', 'p': 'Р', 'c': 'С', 't': 'Т',
	'y': 'У', 'x': 'Х',
}

// RussianLicensePlateRegex регулярное выражение для валидации российских госномеров
// Формат: О111ОО799 (буква + 3 цифры + 2 буквы + 2-3 цифры)
var RussianLicensePlateRegex = regexp.MustCompile(`^[АВЕКМНОРСТУХ]\d{3}[АВЕКМНОРСТУХ]{2}\d{2,3}$`)

// ValidateAndNormalizeLicensePlate валидирует и нормализует российский госномер
// Приводит к формату: заглавные русские буквы + цифры
// Примеры:
//
//	"A123BC77" -> "А123ВС77"
//	"a123bc77" -> "А123ВС77"
//	"О111ОО799" -> "О111ОО799"
//	"o111oo799" -> "О111ОО799"
func ValidateAndNormalizeLicensePlate(licensePlate string) (string, error) {
	if licensePlate == "" {
		return "", ErrEmptyLicensePlate
	}

	// Очищаем от пробелов и дефисов
	cleaned := strings.ReplaceAll(licensePlate, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	// Приводим к верхнему регистру
	upper := strings.ToUpper(cleaned)

	// Заменяем английские буквы на русские
	normalized := replaceEnglishWithRussian(upper)

	// Валидируем формат
	if !RussianLicensePlateRegex.MatchString(normalized) {
		return "", ErrInvalidLicensePlateFormat
	}

	return normalized, nil
}

// replaceEnglishWithRussian заменяет английские буквы на русские в госномере
func replaceEnglishWithRussian(s string) string {
	var result strings.Builder
	for _, r := range s {
		if russian, exists := EnglishToRussian[r]; exists {
			result.WriteRune(russian)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// IsValidLicensePlate проверяет, является ли строка валидным российским госномером
// без нормализации (для быстрой проверки)
func IsValidLicensePlate(licensePlate string) bool {
	if licensePlate == "" {
		return false
	}
	return RussianLicensePlateRegex.MatchString(licensePlate)
}

// NormalizeLicensePlateForSearch нормализует госномер для поиска в базе данных
// Используется в GetUserByCarNumber для поиска по нормализованному номеру
func NormalizeLicensePlateForSearch(licensePlate string) string {
	if licensePlate == "" {
		return ""
	}

	// Очищаем и приводим к верхнему регистру
	cleaned := strings.ReplaceAll(licensePlate, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	upper := strings.ToUpper(cleaned)

	// Заменяем английские буквы на русские
	return replaceEnglishWithRussian(upper)
}

// GetLicensePlateVariants возвращает все возможные варианты госномера для поиска
// Полезно для поиска в БД, где могут быть разные варианты написания
func GetLicensePlateVariants(licensePlate string) []string {
	if licensePlate == "" {
		return []string{}
	}

	normalized, err := ValidateAndNormalizeLicensePlate(licensePlate)
	if err != nil {
		return []string{}
	}

	// Возвращаем нормализованный вариант
	return []string{normalized}
}

// LicensePlateErrors определяет ошибки валидации госномеров
var (
	ErrEmptyLicensePlate         = &LicensePlateError{Code: "EMPTY_LICENSE_PLATE", Message: "Номер автомобиля не может быть пустым"}
	ErrInvalidLicensePlateFormat = &LicensePlateError{Code: "INVALID_FORMAT", Message: "Неверный формат номера автомобиля. Используйте формат: О111ОО799"}
)

// LicensePlateError представляет ошибку валидации госномера
type LicensePlateError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *LicensePlateError) Error() string {
	return e.Message
}

// IsLicensePlateError проверяет, является ли ошибка ошибкой валидации госномера
func IsLicensePlateError(err error) bool {
	_, ok := err.(*LicensePlateError)
	return ok
}
