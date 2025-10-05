package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"modbus-server/internal/config"
	"modbus-server/internal/handlers"
	"modbus-server/internal/service"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Настраиваем Gin
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Создаем Gin роутер
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Создаем сервисы
	modbusService := service.NewModbusService(cfg)

	// Создаем обработчики
	handler := handlers.NewHandler(modbusService)

	// Регистрируем маршруты
	handler.RegisterRoutes(router)

	// Запускаем сервер
	log.Printf("Запуск Modbus сервера на порту %s", cfg.ServerPort)
	log.Printf("Modbus: %v (Host: %s, Port: %d)", cfg.ModbusEnabled, cfg.ModbusHost, cfg.ModbusPort)
	
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
