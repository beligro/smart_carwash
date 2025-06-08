package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения
type Config struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     int
	BackendPort      int
	TelegramToken    string
	TelegramUsername string
	ServerIP         string
	LogLevel         string
	LogPretty        bool
	MetricsEnabled   bool
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	// Загружаем .env файл, если он существует
	_ = godotenv.Load()

	// Получаем значения из переменных окружения
	postgresPort, err := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("неверный формат POSTGRES_PORT: %v", err)
	}

	backendPort, err := strconv.Atoi(getEnv("BACKEND_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("неверный формат BACKEND_PORT: %v", err)
	}

	// Парсим логические значения
	logPretty, _ := strconv.ParseBool(getEnv("LOG_PRETTY", "false"))
	metricsEnabled, _ := strconv.ParseBool(getEnv("METRICS_ENABLED", "true"))

	return &Config{
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "carwash"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     postgresPort,
		BackendPort:      backendPort,
		TelegramToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramUsername: getEnv("TELEGRAM_BOT_USERNAME", ""),
		ServerIP:         getEnv("SERVER_IP", "localhost"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		LogPretty:        logPretty,
		MetricsEnabled:   metricsEnabled,
	}, nil
}

// GetDSN возвращает строку подключения к PostgreSQL
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Moscow",
		c.PostgresHost, c.PostgresUser, c.PostgresPassword, c.PostgresDB, c.PostgresPort)
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
