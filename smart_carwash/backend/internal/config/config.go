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

	// Настройки авторизации
	AdminUsername string
	AdminPassword string
	JWTSecret     string

	// Настройки Tinkoff Kassa
	TinkoffTerminalKey    string
	TinkoffSecretKey      string
	TinkoffBaseURL        string
	PaymentSuccessURL     string
	PaymentFailURL        string
	PaymentWebhookURL     string
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

		// Настройки авторизации
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),

		// Настройки Tinkoff Kassa
		TinkoffTerminalKey: getEnv("TINKOFF_TERMINAL_KEY", ""),
		TinkoffSecretKey:   getEnv("TINKOFF_SECRET_KEY", ""),
		TinkoffBaseURL:     getEnv("TINKOFF_BASE_URL", "https://securepay.tinkoff.ru"),
		PaymentSuccessURL:  getEnv("PAYMENT_SUCCESS_URL", "https://your-domain.com/telegram?payment=success"),
		PaymentFailURL:     getEnv("PAYMENT_FAIL_URL", "https://your-domain.com/telegram?payment=fail"),
		PaymentWebhookURL:  getEnv("PAYMENT_WEBHOOK_URL", "https://your-domain.com/api/v1/payments/webhook"),
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
