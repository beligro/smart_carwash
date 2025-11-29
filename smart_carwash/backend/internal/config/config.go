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
	LimitedAdminUsername string
	LimitedAdminPassword string
	JWTSecret     string

	// Настройки Tinkoff Kassa
	TinkoffTerminalKey string
	TinkoffSecretKey   string
	TinkoffSuccessURL  string
	TinkoffFailURL     string

	// Настройки 1C интеграции
	APIKey1C      string
	CashierUserID string

	// Настройки Modbus
	ModbusEnabled bool
	ModbusHost    string
	ModbusPort    int

	// Настройки Modbus HTTP сервера
	ModbusServerHost string
	ModbusServerPort int

	// Настройки Dahua интеграции
	DahuaWebhookUsername string
	DahuaWebhookPassword string
	DahuaAllowedIPs      string
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

	modbusPort, err := strconv.Atoi(getEnv("MODBUS_PORT", "502"))
	if err != nil {
		return nil, fmt.Errorf("неверный формат MODBUS_PORT: %v", err)
	}

	modbusServerPort, err := strconv.Atoi(getEnv("MODBUS_SERVER_PORT", "8081"))
	if err != nil {
		return nil, fmt.Errorf("неверный формат MODBUS_SERVER_PORT: %v", err)
	}

	modbusEnabled := getEnv("MODBUS_ENABLED", "false") == "true"

	return &Config{
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "carwash"),
		PostgresHost:     getEnv("POSTGRES_HOST", "172.18.0.2"), // Используем статический IP для избежания DNS запросов
		PostgresPort:     postgresPort,
		BackendPort:      backendPort,
		TelegramToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramUsername: getEnv("TELEGRAM_BOT_USERNAME", ""),
		ServerIP:         getEnv("SERVER_IP", "localhost"),

		// Настройки авторизации
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin"),
		LimitedAdminUsername: getEnv("LIMITED_ADMIN_USERNAME", ""),
		LimitedAdminPassword: getEnv("LIMITED_ADMIN_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),

		// Настройки Tinkoff Kassa
		TinkoffTerminalKey: getEnv("TINKOFF_TERMINAL_KEY", ""),
		TinkoffSecretKey:   getEnv("TINKOFF_SECRET_KEY", ""),
		TinkoffSuccessURL:  getEnv("TINKOFF_SUCCESS_URL", "https://t.me/your_bot?startapp=payment_success"),
		TinkoffFailURL:     getEnv("TINKOFF_FAIL_URL", "https://t.me/your_bot?startapp=payment_fail"),

		// Настройки 1C интеграции
		APIKey1C:      getEnv("API_KEY_1C", ""),
		CashierUserID: getEnv("CASHIER_USER_ID", ""),

		// Настройки Modbus
		ModbusEnabled: modbusEnabled,
		ModbusHost:    getEnv("MODBUS_HOST", ""),
		ModbusPort:    modbusPort,

		// Настройки Modbus HTTP сервера
		ModbusServerHost: getEnv("MODBUS_SERVER_HOST", "localhost"),
		ModbusServerPort: modbusServerPort,

		// Настройки Dahua интеграции
		DahuaWebhookUsername: getEnv("DAHUA_WEBHOOK_USERNAME", ""),
		DahuaWebhookPassword: getEnv("DAHUA_WEBHOOK_PASSWORD", ""),
		DahuaAllowedIPs:      getEnv("DAHUA_ALLOWED_IPS", ""),
	}, nil
}

// GetDSN возвращает строку подключения к PostgreSQL
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Moscow connect_timeout=10",
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
