package utils

import (
	"strings"
)

// RussianToEnglish маппинг русских букв на английские для госномеров
var RussianToEnglish = map[rune]rune{
	'А': 'A', 'В': 'B', 'Е': 'E', 'К': 'K', 'М': 'M',
	'Н': 'H', 'О': 'O', 'Р': 'P', 'С': 'C', 'Т': 'T',
	'У': 'Y', 'Х': 'X',
}

// NormalizeLicensePlate нормализует госномер без валидации
// Только переводит русские буквы в английские и убирает дефисы
// Примеры:
//
//	"А123ВС77" -> "A123BC77"
//	"а123вс77" -> "A123BC77"
//	"1234АВ-1" -> "1234AB1" (для Беларуси)
//	"О111ОО799" -> "O111OO799"
func NormalizeLicensePlate(licensePlate string) string {
	if licensePlate == "" {
		return ""
	}

	// Очищаем от пробелов и дефисов
	cleaned := strings.ReplaceAll(licensePlate, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	// Приводим к верхнему регистру
	upper := strings.ToUpper(cleaned)

	// Заменяем русские буквы на английские
	return replaceRussianWithEnglish(upper)
}

// replaceRussianWithEnglish заменяет русские буквы на английские в госномере
func replaceRussianWithEnglish(s string) string {
	var result strings.Builder
	for _, r := range s {
		if english, exists := RussianToEnglish[r]; exists {
			result.WriteRune(english)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// NormalizeLicensePlateForSearch нормализует госномер для поиска в базе данных
// Используется в GetUserByCarNumber для поиска по нормализованному номеру
func NormalizeLicensePlateForSearch(licensePlate string) string {
	return NormalizeLicensePlate(licensePlate)
}
