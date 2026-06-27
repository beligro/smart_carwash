package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof" // Для профилирования и отладки
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"carwash_backend/internal/config"
	authHandlers "carwash_backend/internal/domain/auth/handlers"
	authRepo "carwash_backend/internal/domain/auth/repository"
	authService "carwash_backend/internal/domain/auth/service"
	"carwash_backend/internal/domain/auth/webcodes"
	rusender "carwash_backend/internal/domain/email/rusender"
	carwashStatusHandlers "carwash_backend/internal/domain/carwash_status/handlers"
	carwashStatusRepo "carwash_backend/internal/domain/carwash_status/repository"
	carwashStatusService "carwash_backend/internal/domain/carwash_status/service"
	dahuaHandlers "carwash_backend/internal/domain/dahua/handlers"
	dahuaService "carwash_backend/internal/domain/dahua/service"
	linktokenRepo "carwash_backend/internal/domain/linktoken/repository"
	linktokenService "carwash_backend/internal/domain/linktoken/service"
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
	washboxlogHandlers "carwash_backend/internal/domain/washboxlog/handlers"
	washboxlogRepo "carwash_backend/internal/domain/washboxlog/repository"
	washboxlogService "carwash_backend/internal/domain/washboxlog/service"
	guestHandlers "carwash_backend/internal/domain/guest/handlers"
	webHandlers "carwash_backend/internal/domain/web/handlers"
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
	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

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

	// Применяем миграции (с повторами: Postgres может быть ещё не готов)
	if err := runMigrations(cfg); err != nil {
		log.WithField("error", err).Fatal("Ошибка применения миграций — приложение не запущено")
	}

	// Подключаемся к базе данных с retry механизмом
	db, err := connectToDatabaseWithRetry(cfg)
	if err != nil {
		log.WithField("error", err).Fatal("Ошибка подключения к базе данных после всех попыток")
	}

	// Получаем соединение с базой данных
	sqlDB, err := db.DB()
	if err != nil {
		log.WithField("error", err).Fatal("Ошибка получения соединения с базой данных")
	}

	// Настраиваем пул соединений (оптимизировано для предотвращения исчерпания)
	sqlDB.SetMaxIdleConns(10)                  // Оптимальное количество idle соединений
	sqlDB.SetMaxOpenConns(30)                  // Уменьшено до 30 соединений для стабильности
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Уменьшено для быстрого освобождения
	sqlDB.SetConnMaxIdleTime(30 * time.Second) // Уменьшено для быстрого освобождения

	log.WithFields(logrus.Fields{
		"max_idle_conns":     10,
		"max_open_conns":     30,
		"conn_max_idle_time": "30s",
		"conn_max_lifetime":  "5m",
	}).Info("Database connected successfully with optimized connection pool")

	// SafeDB удален; используем контексты HTTP/фоновых задач напрямую

	// Создаем репозитории
	userRepository := userRepo.NewPostgresRepository(db)
	washboxRepository := washboxRepo.NewPostgresRepository(db)
	sessionRepository := sessionRepo.NewPostgresRepository(db)
	settingsRepository := settingsRepo.NewRepository(db)
	authRepository := authRepo.NewPostgresRepository(db)
	paymentRepository := paymentRepo.NewRepository(db)
	carwashStatusRepository := carwashStatusRepo.NewPostgresRepository(db)
	// Репозиторий логов изменений боксов
	washboxLogRepository := washboxlogRepo.NewPostgresRepository(db)

	// Создаем Tinkoff клиент
	tinkoffClient := paymentTinkoff.NewClient(cfg.TinkoffTerminalKey, cfg.TinkoffSecretKey, cfg.TinkoffSuccessURL, cfg.TinkoffFailURL)

	// Сервис логирования изменений боксов
	washboxLogSvc := washboxlogService.NewService(washboxLogRepository, washboxRepository)

	// Создаем Modbus HTTP адаптер
	modbusAdapter := modbusAdapter.NewModbusAdapter(cfg, db, washboxLogSvc)

	// Создаем сервисы
	userSvc := userService.NewService(userRepository)
	settingsSvc := settingsService.NewService(settingsRepository)
	washboxSvc := washboxService.NewService(washboxRepository, sessionRepository, settingsSvc, db, modbusAdapter, washboxLogSvc)
	// Веб-авторизация: коды на email через RuSender
	webCodeStore := webcodes.NewStore()
	var emailSender *rusender.Client
	if cfg.RusenderApiKey != "" && cfg.RusenderFromEmail != "" {
		emailSender = rusender.NewClient(cfg.RusenderApiKey, cfg.RusenderFromEmail)
	}
	webEmailCodeSender := webcodes.NewWebEmailCodeSender(webCodeStore, emailSender)
	authSvc := authService.NewService(authRepository, cfg, userSvc, webEmailCodeSender)

	// Создаем Modbus service для админских операций
	modbusSvc := modbusService.NewModbusService(db, cfg)

	// Создаем фоновые задачи для кассиров
	backgroundTasks := authService.NewBackgroundTasks(authRepository)

	// Репозиторий и сервис токенов привязки Telegram (для веб-версии и бота)
	linkTokenRepo := linktokenRepo.NewPostgresRepository(db)
	linkTokenSvc := linktokenService.NewService(linkTokenRepo, userRepository, sessionRepository, cfg)

	// Создаем Telegram бота (сбой сети/API — не фатально: API и веб работают без уведомлений в Telegram)
	tgBot, err := telegram.NewBot(userSvc, cfg, linkTokenSvc)
	if err != nil {
		log.WithField("error", err).Error("Telegram бот не инициализирован: уведомления в Telegram и обработка вебхука отключены")
	}
	var sessionTelegram telegram.NotificationService
	if tgBot != nil {
		sessionTelegram = tgBot
	}

	// Создаем сервис сессий с зависимостями (emailSender для уведомлений на email пользователям без Telegram)
	sessionSvc := sessionService.NewService(sessionRepository, washboxSvc, userSvc, sessionTelegram, emailSender, nil, modbusAdapter, settingsSvc, cfg.CashierUserID, appMetrics, db, washboxLogSvc)

	// Создаем сервис платежей с зависимостью от sessionSvc как SessionStatusUpdater и SessionExtensionUpdater
	paymentSvc := paymentService.NewService(paymentRepository, settingsRepository, sessionSvc, sessionSvc, tinkoffClient, cfg.TinkoffTerminalKey, cfg.TinkoffSecretKey, cfg.TinkoffWebSuccessURL, cfg.TinkoffWebFailURL, cfg.TinkoffGuestSuccessURL, cfg.TinkoffGuestFailURL, appMetrics)

	// Обновляем sessionSvc с правильным paymentSvc
	sessionSvc = sessionService.NewService(sessionRepository, washboxSvc, userSvc, sessionTelegram, emailSender, paymentSvc, modbusAdapter, settingsSvc, cfg.CashierUserID, appMetrics, db, washboxLogSvc)

	// Создаем сервис очереди, который зависит от сервисов сессий, боксов и пользователей
	queueSvc := queueService.NewService(sessionSvc, washboxSvc, userSvc, settingsSvc, appMetrics)

	// Устанавливаем вебхук для бота
	if tgBot != nil {
		if err := tgBot.SetWebhook(); err != nil {
			log.WithField("error", err).Warn("Ошибка установки вебхука")
		}
	}

	// Создаем Dahua сервис
	dahuaSvc := dahuaService.NewService(sessionSvc, washboxSvc)

	// Создаем сервис статуса мойки
	carwashStatusSvc := carwashStatusService.NewService(carwashStatusRepository, sessionSvc)

	// Устанавливаем carwashStatusRepo в sessionSvc для проверки статуса при создании сессий
	sessionSvc.SetCarwashStatusRepo(carwashStatusRepository)

	// Создаем обработчики
	userHandler := userHandlers.NewHandler(userSvc)
	washboxHandler := washboxHandlers.NewHandler(washboxSvc)
	sessionHandler := sessionHandlers.NewHandler(sessionSvc, paymentSvc, authSvc, cfg.APIKey1C)
	queueHandler := queueHandlers.NewHandler(queueSvc)
	settingsHandler := settingsHandlers.NewHandler(settingsSvc)
	authHandler := authHandlers.NewHandler(authSvc)
	paymentHandler := paymentHandlers.NewHandler(paymentSvc, authSvc)
	modbusHandler := modbusHandlers.NewHandler(modbusSvc)
	webHandler := webHandlers.NewHandler(userSvc, sessionSvc, linkTokenSvc, paymentSvc)
	guestHandler := guestHandlers.NewHandler(sessionSvc, paymentSvc)
	dahuaHandler := dahuaHandlers.NewHandler(dahuaSvc)
	carwashStatusHandler := carwashStatusHandlers.NewHandler(carwashStatusSvc, authHandler.GetAdminMiddleware())
	// Хендлер истории изменений боксов
	washboxLogHandler := washboxlogHandlers.NewHandler(washboxLogSvc)

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

	// Добавляем endpoint для статистики запросов
	router.GET("/debug/stats", func(c *gin.Context) {
		total, active := middleware.GetRequestStats()
		c.JSON(http.StatusOK, gin.H{
			"total_requests":  total,
			"active_requests": active,
		})
	})

	// Инициализируем маршруты
	api := router.Group("/")
	api.Use(middleware.LoggingMiddleware())
	api.Use(middleware.TimeoutMiddleware(10 * time.Second)) // Таймаут 10 секунд
	{
		// Регистрируем маршруты для каждого домена
		userHandler.RegisterRoutes(api)
		washboxHandler.RegisterRoutes(api, authHandler.GetCleanerMiddleware(), authHandler.GetAdminMiddleware())
		sessionHandler.RegisterRoutes(api)
		queueCashierMiddleware := middleware.CashierMiddleware(authSvc)
		queueHandler.RegisterRoutes(api, queueCashierMiddleware)
		settingsHandler.RegisterRoutes(api)
		authHandler.RegisterRoutes(api)
		paymentHandler.RegisterRoutes(api)
		modbusHandler.RegisterRoutes(api)
		// Гостевое API — без авторизации, доступ по guest_token
		guestGroup := api.Group("/guest")
		guestHandler.RegisterRoutes(guestGroup)

		// Отметка выполненного ТО бокса (кнопка-ссылка в алерте Максу). Доступ по токену.
		api.GET("/maintenance/to-done", func(c *gin.Context) {
			if cfg.MaintenanceToken == "" || c.Query("token") != cfg.MaintenanceToken {
				c.String(http.StatusForbidden, "Доступ запрещён")
				return
			}
			boxNumber := c.Query("box")
			if boxNumber == "" {
				c.String(http.StatusBadRequest, "Не указан бокс")
				return
			}
			// Текущая сумма проданных минут бокса становится новой точкой отсчёта (сброс счётчика)
			var currentMinutes int64
			db.Raw(`SELECT COALESCE(SUM(s.rental_time_minutes + s.extension_time_minutes),0)
				FROM sessions s JOIN wash_boxes w ON w.id = s.box_id
				WHERE w.number = ? AND s.status = 'complete'`, boxNumber).Scan(&currentMinutes)
			// Моточасы на момент отметки (до сброса) — для журнала ТО
			var baselineMinutes int64
			db.Raw(`SELECT COALESCE(baseline_minutes,0) FROM box_maintenance WHERE box_number = ?`, boxNumber).Scan(&baselineMinutes)
			motorHours := (currentMinutes - baselineMinutes) / 60
			if motorHours < 0 {
				motorHours = 0
			}
			res := db.Exec(`UPDATE box_maintenance SET baseline_minutes = ?, last_to_at = NOW(), alerted = false, updated_at = NOW() WHERE box_number = ?`, currentMinutes, boxNumber)
			if res.Error != nil || res.RowsAffected == 0 {
				c.String(http.StatusNotFound, "Бокс не найден в учёте ТО")
				return
			}
			// Журнал ТО: фиксируем отметку из Telegram-кнопки
			db.Exec(`INSERT INTO box_maintenance_log (box_number, performed_by, motor_hours_at_reset, reason, comment)
				VALUES (?,?,?,?,?)`, boxNumber, "Telegram (Макс)", motorHours, "Плановое ТО", "")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><head><meta name='viewport' content='width=device-width,initial-scale=1'></head><body style='font-family:sans-serif;text-align:center;padding:40px;color:#2e7d32'><h2>ТО бокса №"+boxNumber+" отмечено</h2><p style='color:#555'>Счётчик моточасов сброшен. Следующее ТО — через 500 моточасов.</p></body></html>"))
		})

		// Админская мнемосхема ТО аппаратов (учёт моточасов, простои, журнал ТО).
		// Доступ — под админским JWT (как и остальные /admin/* ручки).
		maintAdmin := api.Group("/admin")
		maintAdmin.Use(authHandler.GetAdminMiddleware())
		{
			// GET /admin/box-maintenance — состояние всех боксов учёта ТО
			maintAdmin.GET("/box-maintenance", func(c *gin.Context) {
				const threshold = 450
				const limitMH = 500

				type boxRow struct {
					BoxNumber       int
					TotalMinutes    int64
					BaselineMinutes int64
					LastToAt        time.Time
					Alerted         bool
					BoxStatus       *string
					BoxUpdatedAt    *time.Time
				}
				var rows []boxRow
				db.Raw(`
					SELECT bm.box_number AS box_number,
						COALESCE((SELECT SUM(s.rental_time_minutes + s.extension_time_minutes)
								  FROM sessions s JOIN wash_boxes w2 ON w2.id = s.box_id
								  WHERE w2.number = bm.box_number AND s.status = 'complete'), 0) AS total_minutes,
						bm.baseline_minutes AS baseline_minutes,
						bm.last_to_at AS last_to_at,
						bm.alerted AS alerted,
						w.status AS box_status,
						w.updated_at AS box_updated_at
					FROM box_maintenance bm
					LEFT JOIN wash_boxes w ON w.number = bm.box_number AND w.deleted_at IS NULL
					ORDER BY bm.box_number`).Scan(&rows)

				// Момент входа в текущий сервис (последний переход в maintenance)
				type sinceRow struct {
					BoxNumber int
					EnteredAt time.Time
				}
				var sinceRows []sinceRow
				db.Raw(`SELECT box_number, MAX(created_at) AS entered_at
						FROM washbox_change_logs
						WHERE action = 'status_change' AND new_status = 'maintenance'
						GROUP BY box_number`).Scan(&sinceRows)
				sinceMap := make(map[int]time.Time, len(sinceRows))
				for _, r := range sinceRows {
					sinceMap[r.BoxNumber] = r.EnteredAt
				}

				// Суммарный простой в сервисе за 30 дней (минуты), по журналу смены статусов
				type svcRow struct {
					BoxNumber int
					Minutes   int64
				}
				var svcRows []svcRow
				db.Raw(`
					WITH events AS (
						SELECT box_number, created_at, new_status,
							LEAD(created_at) OVER (PARTITION BY box_number ORDER BY created_at) AS next_at
						FROM washbox_change_logs
						WHERE action = 'status_change' AND new_status IS NOT NULL
					)
					SELECT box_number,
						COALESCE(SUM(EXTRACT(EPOCH FROM (
							LEAST(COALESCE(next_at, NOW()), NOW())
							- GREATEST(created_at, NOW() - INTERVAL '30 days')
						)) / 60), 0)::bigint AS minutes
					FROM events
					WHERE new_status = 'maintenance'
					  AND COALESCE(next_at, NOW()) > NOW() - INTERVAL '30 days'
					GROUP BY box_number`).Scan(&svcRows)
				svcMap := make(map[int]int64, len(svcRows))
				for _, r := range svcRows {
					svcMap[r.BoxNumber] = r.Minutes
				}

				// Журнал ТО — последние 5 записей на бокс
				type histRow struct {
					BoxNumber         int
					PerformedAt       time.Time
					PerformedBy       *string
					MotorHoursAtReset *int
					Reason            *string
					Comment           *string
				}
				var histRows []histRow
				db.Raw(`
					SELECT box_number, performed_at, performed_by, motor_hours_at_reset, reason, comment
					FROM (
						SELECT *, ROW_NUMBER() OVER (PARTITION BY box_number ORDER BY performed_at DESC) AS rn
						FROM box_maintenance_log
					) t
					WHERE rn <= 5
					ORDER BY box_number, performed_at DESC`).Scan(&histRows)
				histMap := make(map[int][]gin.H)
				for _, h := range histRows {
					histMap[h.BoxNumber] = append(histMap[h.BoxNumber], gin.H{
						"performed_at":         h.PerformedAt,
						"performed_by":         h.PerformedBy,
						"motor_hours_at_reset": h.MotorHoursAtReset,
						"reason":               h.Reason,
						"comment":              h.Comment,
					})
				}

				result := make([]gin.H, 0, len(rows))
				for _, r := range rows {
					mh := (r.TotalMinutes - r.BaselineMinutes) / 60
					if mh < 0 {
						mh = 0
					}
					status := "ok"
					if mh >= limitMH {
						status = "overdue"
					} else if mh >= threshold {
						status = "soon"
					}
					percent := int(float64(mh)/float64(limitMH)*100 + 0.5)
					inService := r.BoxStatus != nil && *r.BoxStatus == "maintenance"
					var inServiceSince interface{}
					if inService {
						if t, ok := sinceMap[r.BoxNumber]; ok {
							inServiceSince = t
						} else if r.BoxUpdatedAt != nil {
							inServiceSince = *r.BoxUpdatedAt
						}
					}
					hist := histMap[r.BoxNumber]
					if hist == nil {
						hist = []gin.H{}
					}
					var svc int64
					if v, ok := svcMap[r.BoxNumber]; ok {
						svc = v
					}
					result = append(result, gin.H{
						"box_number":          r.BoxNumber,
						"motor_hours":         mh,
						"last_to_at":          r.LastToAt,
						"threshold":           threshold,
						"limit":               limitMH,
						"percent":             percent,
						"status":              status,
						"alerted":             r.Alerted,
						"in_service":          inService,
						"in_service_since":    inServiceSince,
						"service_minutes_30d": svc,
						"history":             hist,
					})
				}
				c.JSON(http.StatusOK, result)
			})

			// POST /admin/box-maintenance/:box/reset — отметить ТO выполненным (внепланово/планово)
			maintAdmin.POST("/box-maintenance/:box/reset", func(c *gin.Context) {
				boxNum, err := strconv.Atoi(c.Param("box"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный номер бокса"})
					return
				}
				var body struct {
					Reason  string `json:"reason"`
					Comment string `json:"comment"`
				}
				_ = c.ShouldBindJSON(&body)

				var cnt int64
				db.Raw(`SELECT COUNT(*) FROM box_maintenance WHERE box_number = ?`, boxNum).Scan(&cnt)
				if cnt == 0 {
					c.JSON(http.StatusNotFound, gin.H{"error": "бокс не найден в учёте ТО"})
					return
				}

				var totalMinutes, baseline int64
				db.Raw(`SELECT COALESCE(SUM(s.rental_time_minutes + s.extension_time_minutes),0)
						FROM sessions s JOIN wash_boxes w ON w.id = s.box_id
						WHERE w.number = ? AND s.status = 'complete'`, boxNum).Scan(&totalMinutes)
				db.Raw(`SELECT COALESCE(baseline_minutes,0) FROM box_maintenance WHERE box_number = ?`, boxNum).Scan(&baseline)
				mh := (totalMinutes - baseline) / 60
				if mh < 0 {
					mh = 0
				}

				performedBy := "admin"
				if u, ok := c.Get("username"); ok {
					if us, ok2 := u.(string); ok2 && us != "" {
						performedBy = us
					}
				}

				if res := db.Exec(`UPDATE box_maintenance SET baseline_minutes = ?, last_to_at = NOW(), alerted = false, updated_at = NOW() WHERE box_number = ?`, totalMinutes, boxNum); res.Error != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить учёт ТО"})
					return
				}
				db.Exec(`INSERT INTO box_maintenance_log (box_number, performed_by, motor_hours_at_reset, reason, comment)
						VALUES (?,?,?,?,?)`, boxNum, performedBy, mh, body.Reason, body.Comment)

				c.JSON(http.StatusOK, gin.H{
					"status":               "ok",
					"box_number":           boxNum,
					"motor_hours_at_reset": mh,
				})
			})
		}

		// Веб-API для клиентов с JWT (user_id из токена)
		webGroup := api.Group("/web", authHandler.GetWebAuthMiddleware())
		webHandler.RegisterRoutes(webGroup)
		dahuaHandlers.SetupRoutes(api, dahuaHandler)
		carwashStatusHandler.RegisterRoutes(api)
		washboxLogHandler.RegisterRoutes(api, authHandler.GetAdminMiddleware())

		// Вебхук для Telegram бота
		api.POST("/webhook", func(c *gin.Context) {
			if tgBot == nil {
				// 200 — чтобы Telegram не долбил ретраями; обновления теряются, пока бот снова не поднимется
				c.JSON(http.StatusOK, gin.H{"status": "ok", "telegram": "disabled"})
				return
			}
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
			tgBot.ProcessUpdate(update)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// Создаем HTTP сервер с таймаутами
	server := &http.Server{
		Addr:         ":" + os.Getenv("BACKEND_PORT"),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Ожидаем сигнала для завершения
	quit := make(chan os.Signal, 1)
	// Канал для корректного завершения всех фоновых горутин
	done := make(chan struct{})

	// Запускаем pprof сервер для профилирования
	go func() {
		pprofPort := os.Getenv("PPROF_PORT")
		if pprofPort == "" {
			pprofPort = "6060"
		}
		log.WithField("port", pprofPort).Info("🔍 pprof server starting")
		if err := http.ListenAndServe(":"+pprofPort, nil); err != nil {
			log.WithField("error", err).Error("pprof server error")
		}
	}()

	// Запускаем системный мониторинг
	go systemMonitor(done)

	// Запускаем мониторинг БД
	go dbMonitor(db, done)

	// Запускаем polling pending платежей (если включено)
	if os.Getenv("ENABLE_PAYMENT_POLLING") == "true" {
		go paymentPolling(paymentSvc, done)
		logger.Info("Payment polling enabled (interval: 10s)", nil)
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Info("Starting HTTP server", map[string]interface{}{
			"port": os.Getenv("BACKEND_PORT"),
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Ошибка запуска сервера", err)
		}
	}()

	// Запускаем бота в отдельной горутине (long polling), только если инициализация прошла
	if tgBot != nil {
		go func() {
			logger.Info("Starting Telegram bot")
			tgBot.Start()
		}()
	} else {
		logger.Info("Telegram bot long polling не запущен (бот не инициализирован)")
	}

	// Запускаем периодическую задачу для обработки очереди (старт сразу)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := sessionSvc.ProcessQueue(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка обработки очереди")
					}
				}()
			case <-done:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для проверки и завершения истекших сессий (старт через 1 сек)
	go func() {
		time.Sleep(1 * time.Second) // Разносим запуск задач
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if err := sessionSvc.CheckAndCompleteExpiredSessions(ctx); err != nil {
						log.WithField("error", err).Error("Ошибка проверки истекших сессий")
					}
				}()
			case <-done:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для автоматического включения химии каждые 5 секунд (старт через 2 сек)
	go func() {
		time.Sleep(2 * time.Second) // Разносим запуск задач
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := sessionSvc.CheckAndAutoEnableChemistry(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка автоматического включения химии")
					}
				}()
			case <-done:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для проверки и истечения зарезервированных сессий (старт через 3 сек)
	go func() {
		time.Sleep(3 * time.Second) // Разносим запуск задач
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := sessionSvc.CheckAndExpireReservedSessions(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка проверки зарезервированных сессий")
					}
				}()
			case <-done:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для очистки истекших cooldown'ов (старт через 4 сек)
	go func() {
		time.Sleep(4 * time.Second)               // Разносим запуск задач
		ticker := time.NewTicker(5 * time.Second) // Проверяем каждые 5 секунд
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := washboxSvc.CheckCooldownExpired(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка очистки истекших cooldown'ов")
					}
				}()
			case <-done:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для автоматического завершения просроченных уборок (старт через 5 сек)
	go func() {
		time.Sleep(5 * time.Second)                // Разносим запуск задач
		ticker := time.NewTicker(10 * time.Second) // Проверяем каждые 30 секунд
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := washboxSvc.AutoCompleteExpiredCleanings(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка автоматического завершения уборок")
					}
				}()
			case <-done:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для проверки и отправки уведомлений о скором истечении сессий (старт через 6 сек)
	go func() {
		time.Sleep(6 * time.Second) // Разносим запуск задач
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := sessionSvc.CheckAndNotifyExpiringReservedSessions(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка отправки уведомлений о скором истечении сессий")
					}
				}()
			case <-done:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для проверки и отправки уведомлений о скором завершении сессий (старт через 7 сек)
	go func() {
		time.Sleep(7 * time.Second) // Разносим запуск задач
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := sessionSvc.CheckAndNotifyCompletingSessions(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка отправки уведомлений о скором завершении сессий")
					}
				}()
			case <-done:
				return
			}
		}
	}()


	// Запускаем периодическую задачу для отмены брошенных created сессий (старт через 9 сек)
	go func() {
		time.Sleep(9 * time.Second) // Разносим запуск задач
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := sessionSvc.CheckAndCancelAbandonedCreatedSessions(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка отмены брошенных created сессий")
					}
				}()
			case <-done:
				return
			}
		}
	}()

	// Запускаем периодическую задачу для деактивации истекших смен кассиров (старт через 8 сек)
	go func() {
		time.Sleep(8 * time.Second) // Разносим запуск задач
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := backgroundTasks.DeactivateExpiredShifts(ctx2); err != nil {
						log.WithField("error", err).Error("Ошибка деактивации истекших смен кассиров")
					}
				}()
			case <-quit:
				return
			}
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Останавливаем все фоновые горутины
	close(done)

	// Создаем контекст с таймаутом для завершения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем очередь webhook'ов
	paymentSvc.Shutdown()

	// Завершаем сервер
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Ошибка завершения сервера", err)
	}

	logger.Info("Server stopped gracefully")
}

// connectToDatabaseWithRetry подключается к базе данных с повторными попытками
func connectToDatabaseWithRetry(cfg *config.Config) (*gorm.DB, error) {
	maxRetries := 5
	baseDelay := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.WithFields(logrus.Fields{
			"attempt":     attempt,
			"max_retries": maxRetries,
			"host":        cfg.PostgresHost,
		}).Info("Попытка подключения к базе данных")

		db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
		if err == nil {
			logger.WithFields(logrus.Fields{
				"attempt": attempt,
				"host":    cfg.PostgresHost,
			}).Info("Успешное подключение к базе данных")
			return db, nil
		}

		logger.WithFields(logrus.Fields{
			"attempt":     attempt,
			"max_retries": maxRetries,
			"error":       err.Error(),
			"host":        cfg.PostgresHost,
		}).Warn("Ошибка подключения к базе данных")

		if attempt < maxRetries {
			// Экспоненциальная задержка: 1s, 2s, 4s, 8s (2^(attempt-1) секунд)
			delay := time.Duration(1<<uint(attempt-1)) * baseDelay
			logger.WithFields(logrus.Fields{
				"attempt":         attempt,
				"next_attempt_in": delay.String(),
			}).Info("Повторная попытка подключения через")
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("не удалось подключиться к базе данных после %d попыток", maxRetries)
}

// runMigrations применяет миграции к базе данных (с повторами при недоступности БД)
func runMigrations(cfg *config.Config) error {
	logger.Info("Applying database migrations...")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)

	migrationsPath := "./migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = "/app/migrations"
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			return fmt.Errorf("директория с миграциями не найдена: %v", err)
		}
	}

	migrationsPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("ошибка получения пути к миграциям: %v", err)
	}
	// file:// URL должен использовать прямые слэши (важно в Docker/Windows)
	migrationsPath = filepath.ToSlash(migrationsPath)

	logger.Info("Migration path", map[string]interface{}{
		"path": migrationsPath,
	})

	migrationsURL := fmt.Sprintf("file://%s", migrationsPath)

	const maxAttempts = 15
	const delay = 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		m, err := migrate.New(migrationsURL, dsn)
		if err != nil {
			logger.GetLogger().WithFields(logrus.Fields{
				"attempt": attempt,
				"max":     maxAttempts,
				"error":   err,
			}).Warn("Миграции: недоступна БД, повтор через 2 сек...")
			time.Sleep(delay)
			continue
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			m.Close()
			logger.GetLogger().WithFields(logrus.Fields{
				"attempt": attempt,
				"error":   err,
			}).Warn("Миграции: ошибка применения, повтор через 2 сек...")
			time.Sleep(delay)
			continue
		}

		m.Close()
		logger.Info("Database migrations applied successfully")
		return nil
	}

	return fmt.Errorf("не удалось применить миграции после %d попыток", maxAttempts)
}

// systemMonitor запускает системный мониторинг
func systemMonitor(done chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log := logger.GetLogger()
	log.Info("📊 System monitor started")

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			goroutines := runtime.NumGoroutine()

			log.WithFields(logrus.Fields{
				"goroutines":   goroutines,
				"memory_alloc": m.Alloc / 1024 / 1024, // MB
				"memory_sys":   m.Sys / 1024 / 1024,   // MB
				"num_gc":       m.NumGC,
				"time":         time.Now().Format("15:04:05"),
			}).Info("📊 SYSTEM STATS")

			// Алерт если слишком много горутин
			if goroutines > 100 {
				log.WithField("goroutines", goroutines).Warn("⚠️  WARNING: Too many goroutines")
			}

			// Алерт если слишком много памяти
			if m.Alloc > 500*1024*1024 { // 500 MB
				log.WithField("alloc_mb", m.Alloc/1024/1024).Warn("⚠️  WARNING: High memory allocation")
			}

		case <-done:
			log.Info("System monitor stopped")
			return
		}
	}
}

// dbMonitor запускает мониторинг пула соединений БД
func dbMonitor(db *gorm.DB, done chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log := logger.GetLogger()
	log.Info("📊 DB monitor started")

	// Сохраняем настройки пула для сравнения (из main.go:95-98)
	const maxOpenConns = 30
	const maxIdleConns = 10

	for {
		select {
		case <-ticker.C:
			sqlDB, err := db.DB()
			if err != nil {
				log.WithField("error", err).Error("Failed to get DB connection for monitoring")
				continue
			}

			stats := sqlDB.Stats()

			log.WithFields(logrus.Fields{
				"open_connections":     stats.OpenConnections,
				"in_use":               stats.InUse,
				"idle":                 stats.Idle,
				"wait_count":           stats.WaitCount,
				"wait_duration":        stats.WaitDuration.String(),
				"max_open_configured":  maxOpenConns,
				"max_idle_configured":  maxIdleConns,
				"max_idle_closed":      stats.MaxIdleClosed,
				"max_idle_time_closed": stats.MaxIdleTimeClosed,
				"max_lifetime_closed":  stats.MaxLifetimeClosed,
			}).Info("📊 DB POOL STATS")

			// Алерты
			if stats.WaitCount > 0 {
				log.WithFields(logrus.Fields{
					"wait_count":    stats.WaitCount,
					"wait_duration": stats.WaitDuration.String(),
				}).Warn("🚨 ALERT: Queries waiting for DB connection!")
			}

			if stats.InUse >= maxOpenConns-2 {
				log.WithFields(logrus.Fields{
					"in_use": stats.InUse,
					"max":    maxOpenConns,
				}).Warn("⚠️  WARNING: DB pool almost full")
			}

			if stats.Idle >= maxIdleConns-1 {
				log.WithFields(logrus.Fields{
					"idle": stats.Idle,
					"max":  maxIdleConns,
				}).Warn("⚠️  WARNING: DB idle pool almost full")
			}

		case <-done:
			log.Info("DB monitor stopped")
			return
		}
	}
}

// paymentPolling проверяет статус pending платежей через Tinkoff API
func paymentPolling(paymentService paymentService.Service, done chan struct{}) {
	ticker := time.NewTicker(10 * time.Second) // Опрос каждые 10 секунд
	defer ticker.Stop()

	log := logger.GetLogger()
	ctx := context.Background()

	log.Info("Payment polling started (interval: 10s)")

	for {
		select {
		case <-ticker.C:
			if err := paymentService.PollPendingPayments(ctx); err != nil {
				log.WithFields(map[string]interface{}{
					"error": err.Error(),
				}).Error("Payment polling error")
			}

		case <-done:
			log.Info("Payment polling stopped")
			return
		}
	}
}
