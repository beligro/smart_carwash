package utils

import (
	"testing"
)

func TestValidateAndNormalizeLicensePlate(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expected       string
		expectError    bool
		expectedErrMsg string
	}{
		// Валидные номера - русские буквы
		{
			name:        "Valid Russian uppercase",
			input:       "А123ВС77",
			expected:    "А123ВС77",
			expectError: false,
		},
		{
			name:        "Valid Russian lowercase",
			input:       "а123вс77",
			expected:    "А123ВС77",
			expectError: false,
		},
		{
			name:        "Valid Russian mixed case",
			input:       "А123вС77",
			expected:    "А123ВС77",
			expectError: false,
		},
		{
			name:        "Valid Russian with spaces",
			input:       "А 123 ВС 77",
			expected:    "А123ВС77",
			expectError: false,
		},
		{
			name:        "Valid Russian with dashes",
			input:       "А-123-ВС-77",
			expected:    "А123ВС77",
			expectError: false,
		},

		// Валидные номера - английские буквы (должны конвертироваться в русские)
		{
			name:        "Valid English uppercase",
			input:       "A123BC77",
			expected:    "А123ВС77",
			expectError: false,
		},
		{
			name:        "Valid English lowercase",
			input:       "a123bc77",
			expected:    "А123ВС77",
			expectError: false,
		},
		{
			name:        "Valid English mixed case",
			input:       "A123bC77",
			expected:    "А123ВС77",
			expectError: false,
		},
		{
			name:        "Valid English with spaces",
			input:       "A 123 BC 77",
			expected:    "А123ВС77",
			expectError: false,
		},

		// Валидные номера - смешанные языки
		{
			name:        "Valid mixed languages",
			input:       "A123ВС77",
			expected:    "А123ВС77",
			expectError: false,
		},
		{
			name:        "Valid mixed languages reverse",
			input:       "А123BC77",
			expected:    "А123ВС77",
			expectError: false,
		},

		// Валидные номера с 3 цифрами в конце
		{
			name:        "Valid with 3 digits at end",
			input:       "А123ВС777",
			expected:    "А123ВС777",
			expectError: false,
		},
		{
			name:        "Valid English with 3 digits at end",
			input:       "A123BC777",
			expected:    "А123ВС777",
			expectError: false,
		},

		// Невалидные номера
		{
			name:           "Empty string",
			input:          "",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Номер автомобиля не может быть пустым",
		},
		{
			name:           "Invalid format - too short",
			input:          "А123ВС7",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
		{
			name:           "Invalid format - too long",
			input:          "А123ВС7777",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
		{
			name:           "Invalid format - wrong letter position",
			input:          "123АВС77",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
		{
			name:           "Invalid format - wrong digit position",
			input:          "АВС12377",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
		{
			name:           "Invalid format - invalid letter",
			input:          "Z123ВС77",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
		{
			name:           "Invalid format - invalid letter in middle",
			input:          "А123ZС77",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
		{
			name:           "Invalid format - only digits",
			input:          "1234567",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
		{
			name:           "Invalid format - only letters",
			input:          "АВСВСВС",
			expected:       "",
			expectError:    true,
			expectedErrMsg: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateAndNormalizeLicensePlate(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateAndNormalizeLicensePlate() expected error, got nil")
					return
				}
				if err.Error() != tt.expectedErrMsg {
					t.Errorf("ValidateAndNormalizeLicensePlate() error message = %v, want %v", err.Error(), tt.expectedErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateAndNormalizeLicensePlate() unexpected error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ValidateAndNormalizeLicensePlate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidLicensePlate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Валидные номера
		{"Valid Russian", "А123ВС77", true},
		{"Valid English", "A123BC77", true},
		{"Valid with 3 digits", "А123ВС777", true},

		// Невалидные номера
		{"Empty string", "", false},
		{"Too short", "А123ВС7", false},
		{"Too long", "А123ВС7777", false},
		{"Invalid format", "123АВС77", false},
		{"Invalid letter", "Z123ВС77", false},
		{"Only digits", "1234567", false},
		{"Only letters", "АВСВСВС", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidLicensePlate(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidLicensePlate() = %v, want %v", result, tt.expected)
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
		{"Russian uppercase", "А123ВС77", "А123ВС77"},
		{"Russian lowercase", "а123вс77", "А123ВС77"},
		{"English uppercase", "A123BC77", "А123ВС77"},
		{"English lowercase", "a123bc77", "А123ВС77"},
		{"Mixed case", "A123вС77", "А123ВС77"},
		{"With spaces", "A 123 BC 77", "А123ВС77"},
		{"With dashes", "A-123-BC-77", "А123ВС77"},
		{"Empty string", "", ""},
		{"Invalid format", "Z123ВС77", "Z123ВС77"}, // Невалидные номера не изменяются
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

func TestGetLicensePlateVariants(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Valid Russian",
			input:    "А123ВС77",
			expected: []string{"А123ВС77"},
		},
		{
			name:     "Valid English",
			input:    "A123BC77",
			expected: []string{"А123ВС77"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Invalid format",
			input:    "Z123ВС77",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLicensePlateVariants(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("GetLicensePlateVariants() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("GetLicensePlateVariants()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestLicensePlateError(t *testing.T) {
	tests := []struct {
		name     string
		err      *LicensePlateError
		expected string
	}{
		{
			name:     "Empty license plate error",
			err:      ErrEmptyLicensePlate,
			expected: "Номер автомобиля не может быть пустым",
		},
		{
			name:     "Invalid format error",
			err:      ErrInvalidLicensePlateFormat,
			expected: "Неверный формат номера автомобиля. Используйте формат: О111ОО799",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("LicensePlateError.Error() = %v, want %v", tt.err.Error(), tt.expected)
			}
		})
	}
}

func TestIsLicensePlateError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "License plate error",
			err:      ErrEmptyLicensePlate,
			expected: true,
		},
		{
			name:     "Regular error",
			err:      &LicensePlateError{Code: "TEST", Message: "Test error"},
			expected: true,
		},
		{
			name:     "Non-license plate error",
			err:      fmt.Errorf("regular error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLicensePlateError(tt.err)
			if result != tt.expected {
				t.Errorf("IsLicensePlateError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReplaceEnglishWithRussian(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"All English", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "АВСDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"All lowercase English", "abcdefghijklmnopqrstuvwxyz", "АВСDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"Mixed case", "AaBbCc", "ААВВСС"},
		{"Russian letters", "АВЕКМНОРСТУХ", "АВЕКМНОРСТУХ"},
		{"Numbers and symbols", "A123B-C", "А123В-С"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceEnglishWithRussian(tt.input)
			if result != tt.expected {
				t.Errorf("replaceEnglishWithRussian() = %v, want %v", result, tt.expected)
			}
		})
	}
}
