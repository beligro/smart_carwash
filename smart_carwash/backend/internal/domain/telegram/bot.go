package telegram

import (
	"carwash_backend/internal/logger"
	"fmt"
	"strings"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/user/models"
	"carwash_backend/internal/domain/user/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// NotificationType тип уведомления
type NotificationType string

const (
	// NotificationTypeSessionExpiringSoon уведомление о скором истечении сессии
	NotificationTypeSessionExpiringSoon NotificationType = "session_expiring_soon"
	// NotificationTypeSessionCompletingSoon уведомление о скором завершении сессии
	NotificationTypeSessionCompletingSoon NotificationType = "session_completing_soon"
	// NotificationTypeBoxAssigned уведомление о назначении бокса
	NotificationTypeBoxAssigned NotificationType = "box_assigned"
)

// NotificationService интерфейс для отправки уведомлений
type NotificationService interface {
	SendSessionNotification(telegramID int64, notificationType NotificationType) error
	SendBoxAssignmentNotification(telegramID int64, boxNumber int) error
	SendSessionReassignmentNotification(telegramID int64, serviceType string) error
}

// Bot структура для работы с Telegram ботом
type Bot struct {
	bot     *tgbotapi.BotAPI
	service service.Service
	config  *config.Config
}

// NewBot создает новый экземпляр Bot
func NewBot(service service.Service, config *config.Config) (*Bot, error) {
	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %v", err)
	}

	// Устанавливаем режим отладки
	bot.Debug = false

	logger.Printf("Авторизован как %s", bot.Self.UserName)

	return &Bot{
		bot:     bot,
		service: service,
		config:  config,
	}, nil
}

// Start запускает бота в режиме long polling
func (b *Bot) Start() {
	logger.Info("Бот запущен в режиме long polling")

	// Удаляем вебхук перед использованием long polling
	_, err := b.bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		logger.Printf("Ошибка удаления вебхука: %v", err)
		return
	}

	logger.Info("Вебхук удален, бот работает в режиме long polling")

	// Настраиваем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Получаем канал обновлений
	updates := b.bot.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		// Обрабатываем сообщения
		if update.Message != nil {
			b.handleMessage(update.Message)
		}
	}
}

// ProcessUpdate обрабатывает обновления от вебхука
func (b *Bot) ProcessUpdate(update tgbotapi.Update) {
	// Обрабатываем сообщения
	if update.Message != nil {
		b.handleMessage(update.Message)
	}
}

// SetWebhook устанавливает вебхук для бота
func (b *Bot) SetWebhook() error {
	// Формируем URL вебхука
	webhookURL := fmt.Sprintf("https://%s/webhook", b.config.ServerIP)

	// Создаем конфигурацию вебхука
	webhookConfig, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("ошибка установки вебхука: %v", err)
	}

	// Устанавливаем вебхук
	_, err = b.bot.Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("ошибка установки вебхука: %v", err)
	}

	logger.Printf("Вебхук установлен на %s", webhookURL)
	return nil
}

// handleMessage обрабатывает сообщения
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// Обрабатываем команду /start
	if message.IsCommand() && message.Command() == "start" {
		b.handleStartCommand(message)
		return
	}

	// Отправляем сообщение с помощью
	b.sendHelpMessage(message.Chat.ID)
}

// handleStartCommand обрабатывает команду /start
func (b *Bot) handleStartCommand(message *tgbotapi.Message) {
	// Создаем пользователя
	_, err := b.service.CreateUser(&models.CreateUserRequest{
		TelegramID: message.From.ID,
		Username:   message.From.UserName,
		FirstName:  message.From.FirstName,
		LastName:   message.From.LastName,
	})

	if err != nil {
		logger.Printf("Ошибка создания пользователя: %v", err)
		b.sendErrorMessage(message.Chat.ID)
		return
	}

	// Формируем приветственное сообщение
	messageText := "Помыть машину/записаться в очередь."

	// Создаем inline клавиатуру с кнопкой для mini app
	// Используем специальный формат URL для Telegram Mini App
	miniAppURL := fmt.Sprintf("https://t.me/%s?startapp=carwash", b.config.TelegramUsername)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{
				Text: "→ НАЖМИТЕ СЮДА",
				URL:  &miniAppURL,
			},
		),
	)

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(message.Chat.ID, messageText)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = b.bot.Send(msg)
	if err != nil {
		logger.Printf("Ошибка отправки сообщения: %v", err)
	}
}

// sendHelpMessage отправляет сообщение с помощью
func (b *Bot) sendHelpMessage(chatID int64) {
	// Формируем сообщение с помощью
	var messageText strings.Builder
	messageText.WriteString("Доступные команды:\n\n")
	messageText.WriteString("/start - Начать работу с ботом")

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ParseMode = "HTML"

	_, err := b.bot.Send(msg)
	if err != nil {
		logger.Printf("Ошибка отправки сообщения: %v", err)
	}
}

// sendErrorMessage отправляет сообщение об ошибке
func (b *Bot) sendErrorMessage(chatID int64) {
	// Отправляем сообщение об ошибке
	msg := tgbotapi.NewMessage(chatID, "Произошла ошибка. Пожалуйста, попробуйте позже.")
	_, err := b.bot.Send(msg)
	if err != nil {
		logger.Printf("Ошибка отправки сообщения: %v", err)
	}
}

// SendSessionNotification отправляет уведомление о сессии
func (b *Bot) SendSessionNotification(telegramID int64, notificationType NotificationType) error {
	var messageText string

	switch notificationType {
	case NotificationTypeSessionExpiringSoon:
		messageText = "Внимание! Через 1 минуту истечет время ожидания начала мойки. Пожалуйста, начните мойку, иначе ваша сессия будет отменена. Деньги возвращены, очередь аннулирована."
	case NotificationTypeSessionCompletingSoon:
		messageText = "⚠️ Внимание! Через 5 минут завершится время мойки. Самое время продлить оплаченное время или поторопиться."
	default:
		return fmt.Errorf("неизвестный тип уведомления: %s", notificationType)
	}

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(telegramID, messageText)
	msg.ParseMode = "HTML"

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки уведомления: %v", err)
	}

	return nil
}

// SendBoxAssignmentNotification отправляет уведомление о назначении бокса
func (b *Bot) SendBoxAssignmentNotification(telegramID int64, boxNumber int) error {
	messageText := fmt.Sprintf("Вам назначен бокс №%d! Добро пожаловать. Будьте осторожны и заезжайте в бокс! У вас 3 минуты для начала мойки в мини приложении.", boxNumber)

	// Создаем inline клавиатуру с кнопкой для mini app
	// Используем специальный формат URL для Telegram Mini App
	miniAppURL := fmt.Sprintf("https://t.me/%s?startapp=carwash", b.config.TelegramUsername)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{
				Text: "→ НАЖМИТЕ СЮДА",
				URL:  &miniAppURL,
			},
		),
	)

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(telegramID, messageText)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки уведомления о назначении бокса: %v", err)
	}

	return nil
}

// SendSessionReassignmentNotification отправляет уведомление о переназначении сессии
func (b *Bot) SendSessionReassignmentNotification(telegramID int64, serviceType string) error {
	var serviceText string
	switch serviceType {
	case "wash":
		serviceText = "мойки"
	case "air_dry":
		serviceText = "обдува"
	case "vacuum":
		serviceText = "пылесоса"
	default:
		serviceText = "услуги"
	}

	messageText := fmt.Sprintf("🔄 Ваша сессия %s была переназначена на другой бокс. Пожалуйста, ожидайте уведомления о новом боксе.", serviceText)

	// Отправляем сообщение
	msg := tgbotapi.NewMessage(telegramID, messageText)
	msg.ParseMode = "HTML"

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки уведомления о переназначении сессии: %v", err)
	}

	return nil
}
