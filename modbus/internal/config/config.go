package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config содержит конфигурацию modbus сервера
type Config struct {
	// HTTP сервер
	ServerPort string

	// Modbus TCP
	ModbusEnabled bool
	ModbusHost    string
	ModbusPort    int

	// Логирование
	LogLevel string
}

// LoadConfig загружает конфигурацию из config.env файла
func LoadConfig() (*Config, error) {
	// Загружаем config.env файл если он существует
	if err := godotenv.Load("config.env"); err != nil {
		// Игнорируем ошибку если файл не найден
	}

	cfg := &Config{
		ServerPort:    getEnv("SERVER_PORT", "8081"),
		ModbusEnabled: getEnvBool("MODBUS_ENABLED", true),
		ModbusHost:    getEnv("MODBUS_HOST", "localhost"),
		ModbusPort:    getEnvInt("MODBUS_PORT", 502),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt получает переменную окружения как int или возвращает значение по умолчанию
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool получает переменную окружения как bool или возвращает значение по умолчанию
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
