package utils

import (
	"testing"
)

func TestNormalizeLicensePlate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Валидные номера - русские буквы (должны конвертироваться в английские)
		{
			name:     "Valid Russian uppercase",
			input:    "А123ВС77",
			expected: "A123BC77",
		},
		{
			name:     "Valid Russian lowercase",
			input:    "а123вс77",
			expected: "A123BC77",
		},
		{
			name:     "Valid Russian mixed case",
			input:    "А123вС77",
			expected: "A123BC77",
		},
		{
			name:     "Valid Russian with spaces",
			input:    "А 123 ВС 77",
			expected: "A123BC77",
		},
		{
			name:     "Valid Russian with dashes",
			input:    "А-123-ВС-77",
			expected: "A123BC77",
		},

		// Валидные номера - английские буквы (остаются как есть)
		{
			name:     "Valid English uppercase",
			input:    "A123BC77",
			expected: "A123BC77",
		},
		{
			name:     "Valid English lowercase",
			input:    "a123bc77",
			expected: "A123BC77",
		},
		{
			name:     "Valid English mixed case",
			input:    "A123bC77",
			expected: "A123BC77",
		},
		{
			name:     "Valid English with spaces",
			input:    "A 123 BC 77",
			expected: "A123BC77",
		},

		// Валидные номера - смешанные языки
		{
			name:     "Valid mixed languages",
			input:    "A123ВС77",
			expected: "A123BC77",
		},
		{
			name:     "Valid mixed languages reverse",
			input:    "А123BC77",
			expected: "A123BC77",
		},

		// Валидные номера с 3 цифрами в конце
		{
			name:     "Valid with 3 digits at end",
			input:    "А123ВС777",
			expected: "A123BC777",
		},
		{
			name:     "Valid English with 3 digits at end",
			input:    "A123BC777",
			expected: "A123BC777",
		},

		// Нормализация для Беларуси (убираем дефис)
		{
			name:     "Belarus with dash",
			input:    "1234АВ-1",
			expected: "1234AB1",
		},
		{
			name:     "Belarus without dash",
			input:    "1234АВ1",
			expected: "1234AB1",
		},

		// Пустые строки
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},

		// Невалидные номера (нормализация все равно работает)
		{
			name:     "Invalid format - too short",
			input:    "А123ВС7",
			expected: "A123BC7",
		},
		{
			name:     "Invalid format - too long",
			input:    "А123ВС7777",
			expected: "A123BC7777",
		},
		{
			name:     "Invalid format - wrong letter position",
			input:    "123АВС77",
			expected: "123ABC77",
		},
		{
			name:     "Invalid format - wrong digit position",
			input:    "АВС12377",
			expected: "ABC12377",
		},
		{
			name:     "Invalid format - invalid letter",
			input:    "Z123ВС77",
			expected: "Z123BC77",
		},
		{
			name:     "Invalid format - invalid letter in middle",
			input:    "А123ZС77",
			expected: "A123ZC77",
		},
		{
			name:     "Invalid format - only digits",
			input:    "1234567",
			expected: "1234567",
		},
		{
			name:     "Invalid format - only letters",
			input:    "АВСВСВС",
			expected: "ABCBCBC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeLicensePlate(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeLicensePlate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeLicensePlateForSearch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Нормализация для поиска
		{"Russian uppercase", "А123ВС77", "A123BC77"},
		{"Russian lowercase", "а123вс77", "A123BC77"},
		{"English uppercase", "A123BC77", "A123BC77"},
		{"English lowercase", "a123bc77", "A123BC77"},
		{"Mixed case", "A123вС77", "A123BC77"},
		{"With spaces", "A 123 BC 77", "A123BC77"},
		{"With dashes", "A-123-BC-77", "A123BC77"},
		{"Empty string", "", ""},
		{"Invalid format", "Z123ВС77", "Z123BC77"}, // Невалидные номера не изменяются
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeLicensePlateForSearch(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeLicensePlateForSearch() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReplaceRussianWithEnglish(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"All Russian", "АВЕКМНОРСТУХ", "ABEKMHOPCTYX"},
		{"All lowercase Russian", "авекмнорстух", "ABEKMHOPCTYX"},
		{"Mixed case", "АаВвСс", "AaBbCc"},
		{"English letters", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"Numbers and symbols", "А123В-С", "A123B-C"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceRussianWithEnglish(tt.input)
			if result != tt.expected {
				t.Errorf("replaceRussianWithEnglish() = %v, want %v", result, tt.expected)
			}
		})
	}
}
