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
	user, err := b.service.CreateUser(&models.CreateUserRequest{
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
	var messageText strings.Builder
	messageText.WriteString(fmt.Sprintf("Привет, %s!\n\n", user.FirstName))
	messageText.WriteString("Добро пожаловать в бота умной автомойки.")

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(message.Chat.ID, messageText.String())
	msg.ParseMode = "HTML"

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
		messageText = "⚠️ Внимание! Через 1 минуту истечет время ожидания начала мойки. Пожалуйста, начните мойку, иначе ваша сессия будет отменена."
	case NotificationTypeSessionCompletingSoon:
		messageText = "⚠️ Внимание! Через 1 минуту завершится время мойки. Пожалуйста, завершите мойку."
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
	messageText := fmt.Sprintf("Вам назначен бокс №%d! Пожалуйста, подъедьте к указанному боксу и начните мойку.", boxNumber)

	// Отправляем сообщение
	msg := tgbotapi.NewMessage(telegramID, messageText)
	msg.ParseMode = "HTML"

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка отправки уведомления о назначении бокса: %v", err)
	}

	return nil
}
