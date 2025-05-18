package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"carwash_backend/internal/config"
	"carwash_backend/internal/handlers"
	"carwash_backend/internal/repository"
	"carwash_backend/internal/service"
	"carwash_backend/internal/telegram"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Применяем миграции
	if err := runMigrations(cfg); err != nil {
		log.Printf("Ошибка применения миграций: %v", err)
	}

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Получаем соединение с базой данных
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Ошибка получения соединения с базой данных: %v", err)
	}

	// Настраиваем пул соединений
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Создаем репозиторий
	repo := repository.NewPostgresRepository(db)

	// Создаем сервис
	svc := service.NewService(repo)

	// Создаем Telegram бота
	bot, err := telegram.NewBot(svc, cfg)
	if err != nil {
		log.Fatalf("Ошибка создания Telegram бота: %v", err)
	}

	// Устанавливаем вебхук для бота
	if err := bot.SetWebhook(); err != nil {
		log.Printf("Ошибка установки вебхука: %v", err)
	}

	// Создаем обработчики
	handler := handlers.NewHandler(svc, bot)

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

	// Ожидаем сигнала для завершения
	quit := make(chan os.Signal, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Запускаем бота в отдельной горутине
	go bot.Start()

	// Запускаем периодическую задачу для обработки очереди
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Запуск обработки очереди...")
				if err := svc.ProcessQueue(); err != nil {
					log.Printf("Ошибка обработки очереди: %v", err)
				} else {
					log.Println("Обработка очереди завершена успешно")
				}
			case <-quit:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для проверки и завершения истекших сессий
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Запуск проверки истекших сессий...")
				if err := svc.CheckAndCompleteExpiredSessions(); err != nil {
					log.Printf("Ошибка проверки истекших сессий: %v", err)
				} else {
					log.Println("Проверка истекших сессий завершена успешно")
				}
			case <-quit:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для проверки и истечения зарезервированных сессий
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Запуск проверки зарезервированных сессий...")
				if err := svc.CheckAndExpireReservedSessions(); err != nil {
					log.Printf("Ошибка проверки зарезервированных сессий: %v", err)
				} else {
					log.Println("Проверка зарезервированных сессий завершена успешно")
				}
			case <-quit:
				return
			}
		}
	}()
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

// runMigrations применяет миграции к базе данных
func runMigrations(cfg *config.Config) error {
	log.Println("Применение миграций...")

	// Формируем DSN для миграций
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)

	// Определяем путь к директории с миграциями
	migrationsPath := "./migrations"

	// Проверяем существование директории
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// Если директория не существует, пробуем другой путь
		migrationsPath = "/app/migrations"
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			return fmt.Errorf("директория с миграциями не найдена: %v", err)
		}
	}

	// Получаем абсолютный путь к директории с миграциями
	migrationsPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("ошибка получения пути к миграциям: %v", err)
	}

	log.Printf("Путь к миграциям: %s", migrationsPath)

	// Создаем URL для миграций
	migrationsURL := fmt.Sprintf("file://%s", migrationsPath)

	// Создаем экземпляр migrate
	m, err := migrate.New(migrationsURL, dsn)
	if err != nil {
		return fmt.Errorf("ошибка создания экземпляра migrate: %v", err)
	}

	// Применяем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций: %v", err)
	}

	log.Println("Миграции успешно применены")
	return nil
}
