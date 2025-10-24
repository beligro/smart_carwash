package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"carwash_backend/internal/config"
	authHandlers "carwash_backend/internal/domain/auth/handlers"
	authRepo "carwash_backend/internal/domain/auth/repository"
	authService "carwash_backend/internal/domain/auth/service"
	dahuaHandlers "carwash_backend/internal/domain/dahua/handlers"
	dahuaService "carwash_backend/internal/domain/dahua/service"
	modbusAdapter "carwash_backend/internal/domain/modbus/adapter"
	modbusHandlers "carwash_backend/internal/domain/modbus/handlers"
	modbusService "carwash_backend/internal/domain/modbus/service"
	paymentHandlers "carwash_backend/internal/domain/payment/handlers"
	paymentRepo "carwash_backend/internal/domain/payment/repository"
	paymentService "carwash_backend/internal/domain/payment/service"
	paymentTinkoff "carwash_backend/internal/domain/payment/tinkoff"
	queueHandlers "carwash_backend/internal/domain/queue/handlers"
	queueService "carwash_backend/internal/domain/queue/service"
	sessionHandlers "carwash_backend/internal/domain/session/handlers"
	sessionRepo "carwash_backend/internal/domain/session/repository"
	sessionService "carwash_backend/internal/domain/session/service"
	settingsHandlers "carwash_backend/internal/domain/settings/handlers"
	settingsRepo "carwash_backend/internal/domain/settings/repository"
	settingsService "carwash_backend/internal/domain/settings/service"
	"carwash_backend/internal/domain/telegram"
	userHandlers "carwash_backend/internal/domain/user/handlers"
	userRepo "carwash_backend/internal/domain/user/repository"
	userService "carwash_backend/internal/domain/user/service"
	washboxHandlers "carwash_backend/internal/domain/washbox/handlers"
	washboxRepo "carwash_backend/internal/domain/washbox/repository"
	washboxService "carwash_backend/internal/domain/washbox/service"
	"carwash_backend/internal/logger"
	"carwash_backend/internal/metrics"
	"carwash_backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Инициализируем структурированный логгер
	logger.Init()
	log := logger.GetLogger()
	log.Info("Starting Smart Carwash Backend")

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Ошибка загрузки конфигурации", err)
	}

	// Инициализируем метрики
	appMetrics := metrics.NewMetrics()
	log.Info("Metrics initialized")

	// Применяем миграции
	if err := runMigrations(cfg); err != nil {
		log.WithField("error", err).Error("Ошибка применения миграций")
	}

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.WithField("error", err).Fatal("Ошибка подключения к базе данных")
	}

	// Получаем соединение с базой данных
	sqlDB, err := db.DB()
	if err != nil {
		log.WithField("error", err).Fatal("Ошибка получения соединения с базой данных")
	}

	// Настраиваем пул соединений
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.WithFields(logrus.Fields{
		"max_idle_conns": 10,
		"max_open_conns": 100,
	}).Info("Database connected successfully")

	// Создаем репозитории
	userRepository := userRepo.NewPostgresRepository(db)
	washboxRepository := washboxRepo.NewPostgresRepository(db)
	sessionRepository := sessionRepo.NewPostgresRepository(db)
	settingsRepository := settingsRepo.NewRepository(db)
	authRepository := authRepo.NewPostgresRepository(db)
	paymentRepository := paymentRepo.NewRepository(db)

	// Создаем Tinkoff клиент
	tinkoffClient := paymentTinkoff.NewClient(cfg.TinkoffTerminalKey, cfg.TinkoffSecretKey, cfg.TinkoffSuccessURL, cfg.TinkoffFailURL)

	// Создаем сервисы
	userSvc := userService.NewService(userRepository)
	settingsSvc := settingsService.NewService(settingsRepository)
	washboxSvc := washboxService.NewService(washboxRepository, settingsSvc, db)
	authSvc := authService.NewService(authRepository, cfg)

	// Создаем Modbus HTTP адаптер
	modbusAdapter := modbusAdapter.NewModbusAdapter(cfg, db)

	// Создаем Modbus service для админских операций
	modbusSvc := modbusService.NewModbusService(db, cfg)

	// Создаем фоновые задачи для кассиров
	backgroundTasks := authService.NewBackgroundTasks(authRepository)

	// Создаем Telegram бота
	bot, err := telegram.NewBot(userSvc, cfg)
	if err != nil {
		log.WithField("error", err).Fatal("Ошибка создания Telegram бота")
	}

	// Создаем сервис сессий с зависимостями
	sessionSvc := sessionService.NewService(sessionRepository, washboxSvc, userSvc, bot, nil, modbusAdapter, settingsSvc, cfg.CashierUserID, appMetrics) // paymentSvc будет nil пока

	// Создаем сервис платежей с зависимостью от sessionSvc как SessionStatusUpdater и SessionExtensionUpdater
	paymentSvc := paymentService.NewService(paymentRepository, settingsRepository, sessionSvc, sessionSvc, tinkoffClient, cfg.TinkoffTerminalKey, cfg.TinkoffSecretKey, appMetrics)

	// Обновляем sessionSvc с правильным paymentSvc
	sessionSvc = sessionService.NewService(sessionRepository, washboxSvc, userSvc, bot, paymentSvc, modbusAdapter, settingsSvc, cfg.CashierUserID, appMetrics)

	// Создаем сервис очереди, который зависит от сервисов сессий, боксов и пользователей
	queueSvc := queueService.NewService(sessionSvc, washboxSvc, userSvc, appMetrics)

	// Устанавливаем вебхук для бота
	if err := bot.SetWebhook(); err != nil {
		log.WithField("error", err).Warn("Ошибка установки вебхука")
	}

	// Создаем Dahua сервис
	dahuaSvc := dahuaService.NewService(userSvc, sessionSvc)

	// Создаем обработчики
	userHandler := userHandlers.NewHandler(userSvc)
	washboxHandler := washboxHandlers.NewHandler(washboxSvc)
	sessionHandler := sessionHandlers.NewHandler(sessionSvc, paymentSvc, authSvc, cfg.APIKey1C)
	queueHandler := queueHandlers.NewHandler(queueSvc)
	settingsHandler := settingsHandlers.NewHandler(settingsSvc)
	authHandler := authHandlers.NewHandler(authSvc)
	paymentHandler := paymentHandlers.NewHandler(paymentSvc, authSvc)
	modbusHandler := modbusHandlers.NewHandler(modbusSvc)
	dahuaHandler := dahuaHandlers.NewHandler(dahuaSvc)

	// Создаем роутер
	router := gin.Default()

	// Добавляем middleware для метрик
	router.Use(appMetrics.PrometheusMiddleware())

	// Настраиваем CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Добавляем endpoint для метрик
	router.GET("/metrics", appMetrics.MetricsHandler())

	// Инициализируем маршруты
	api := router.Group("/")
	api.Use(middleware.LoggingMiddleware())
	{
		// Регистрируем маршруты для каждого домена
		userHandler.RegisterRoutes(api)
		washboxHandler.RegisterRoutes(api, authHandler.GetCleanerMiddleware())
		sessionHandler.RegisterRoutes(api)
		queueHandler.RegisterRoutes(api)
		settingsHandler.RegisterRoutes(api)
		authHandler.RegisterRoutes(api)
		paymentHandler.RegisterRoutes(api)
		modbusHandler.RegisterRoutes(api)
		dahuaHandlers.SetupRoutes(api, dahuaHandler)

		// Вебхук для Telegram бота
		api.POST("/webhook", func(c *gin.Context) {
			// Читаем тело запроса
			body, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось прочитать тело запроса"})
				return
			}

			// Парсим обновление
			var update tgbotapi.Update
			if err := json.Unmarshal(body, &update); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось распарсить обновление"})
				return
			}

			// Обрабатываем обновление
			bot.ProcessUpdate(update)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:    ":" + os.Getenv("BACKEND_PORT"),
		Handler: router,
	}

	// Ожидаем сигнала для завершения
	quit := make(chan os.Signal, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Info("Starting HTTP server", map[string]interface{}{
			"port": os.Getenv("BACKEND_PORT"),
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Ошибка запуска сервера", err)
		}
	}()

	// Запускаем бота в отдельной горутине
	go func() {
		logger.Info("Starting Telegram bot")
		bot.Start()
	}()

	// Запускаем периодическую задачу для обработки очереди
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := sessionSvc.ProcessQueue(); err != nil {
					log.WithField("error", err).Error("Ошибка обработки очереди")
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
				if err := sessionSvc.CheckAndCompleteExpiredSessions(); err != nil {
					log.WithField("error", err).Error("Ошибка проверки истекших сессий")
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
				if err := sessionSvc.CheckAndExpireReservedSessions(); err != nil {
					log.WithField("error", err).Error("Ошибка проверки зарезервированных сессий")
				}
			case <-quit:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для очистки истекших cooldown'ов
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Проверяем каждую минуту
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := washboxSvc.CheckCooldownExpired(); err != nil {
					log.WithField("error", err).Error("Ошибка очистки истекших cooldown'ов")
				}
			case <-quit:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для автоматического завершения просроченных уборок
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Проверяем каждые 30 секунд
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := washboxSvc.AutoCompleteExpiredCleanings(); err != nil {
					log.WithField("error", err).Error("Ошибка автоматического завершения уборок")
				}
			case <-quit:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для проверки и отправки уведомлений о скором истечении сессий
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := sessionSvc.CheckAndNotifyExpiringReservedSessions(); err != nil {
					log.WithField("error", err).Error("Ошибка отправки уведомлений о скором истечении сессий")
				}
			case <-quit:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для проверки и отправки уведомлений о скором завершении сессий
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := sessionSvc.CheckAndNotifyCompletingSessions(); err != nil {
					log.WithField("error", err).Error("Ошибка отправки уведомлений о скором завершении сессий")
				}
			case <-quit:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для деактивации истекших смен кассиров
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := backgroundTasks.DeactivateExpiredShifts(); err != nil {
					log.WithField("error", err).Error("Ошибка деактивации истекших смен кассиров")
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
		logger.Fatal("Ошибка завершения сервера", err)
	}

	logger.Info("Server stopped gracefully")
}

// runMigrations применяет миграции к базе данных
func runMigrations(cfg *config.Config) error {
	logger.Info("Applying database migrations...")

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

	logger.Info("Migration path", map[string]interface{}{
		"path": migrationsPath,
	})

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

	logger.Info("Database migrations applied successfully")
	return nil
}
