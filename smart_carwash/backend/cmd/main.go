package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"carwash_backend/internal/config"
	"carwash_backend/internal/handlers"
	"carwash_backend/internal/repository"
	"carwash_backend/internal/service"
	"carwash_backend/internal/telegram"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Подключаемся к базе данных
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Создаем репозиторий
	repo := repository.NewPostgresRepository(db)

	// Создаем сервис
	svc := service.NewService(repo)

	// Создаем обработчики
	handler := handlers.NewHandler(svc)

	// Создаем Telegram бота
	bot, err := telegram.NewBot(svc, cfg)
	if err != nil {
		log.Fatalf("Ошибка создания Telegram бота: %v", err)
	}

	// Устанавливаем вебхук для бота
	if err := bot.SetWebhook(); err != nil {
		log.Printf("Ошибка установки вебхука: %v", err)
	}

	// Создаем роутер
	router := gin.Default()

	// Настраиваем CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Инициализируем маршруты
	handler.InitRoutes(router)

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:    ":" + os.Getenv("BACKEND_PORT"),
		Handler: router,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Запускаем бота в отдельной горутине
	go bot.Start()

	// Ожидаем сигнала для завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Создаем контекст с таймаутом для завершения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Завершаем сервер
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка завершения сервера: %v", err)
	}

	log.Println("Сервер остановлен")
}

// connectDB подключается к базе данных
func connectDB(cfg *config.Config) (*gorm.DB, error) {
	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Получаем соединение с базой данных
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Настраиваем пул соединений
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
