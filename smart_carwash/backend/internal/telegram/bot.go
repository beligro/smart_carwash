package telegram

import (
	"fmt"
	"log"
	"strings"

	"carwash_backend/internal/config"
	"carwash_backend/internal/models"
	"carwash_backend/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

	log.Printf("Авторизован как %s", bot.Self.UserName)

	return &Bot{
		bot:     bot,
		service: service,
		config:  config,
	}, nil
}

// Start запускает бота в режиме long polling
func (b *Bot) Start() {
	log.Println("Бот запущен в режиме long polling")

	// Удаляем вебхук перед использованием long polling
	_, err := b.bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Printf("Ошибка удаления вебхука: %v", err)
		return
	}

	log.Println("Вебхук удален, бот работает в режиме long polling")

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

	log.Printf("Вебхук установлен на %s", webhookURL)
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
		log.Printf("Ошибка создания пользователя: %v", err)
		b.sendErrorMessage(message.Chat.ID)
		return
	}

	// Формируем приветственное сообщение
	var messageText strings.Builder
	messageText.WriteString(fmt.Sprintf("Привет, %s!\n\n", user.FirstName))
	messageText.WriteString("Добро пожаловать в бота умной автомойки.\n\n")
	messageText.WriteString("Нажмите на кнопку ниже, чтобы открыть приложение.")

	// Создаем клавиатуру с кнопкой для открытия Mini App
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Открыть приложение", fmt.Sprintf("https://%s", b.config.ServerIP)),
		),
	)

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(message.Chat.ID, messageText.String())
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "HTML"

	_, err = b.bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

// sendHelpMessage отправляет сообщение с помощью
func (b *Bot) sendHelpMessage(chatID int64) {
	// Формируем сообщение с помощью
	var messageText strings.Builder
	messageText.WriteString("Доступные команды:\n\n")
	messageText.WriteString("/start - Начать работу с ботом\n\n")
	messageText.WriteString("Нажмите на кнопку ниже, чтобы открыть приложение.")

	// Создаем клавиатуру с кнопкой для открытия Mini App
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Открыть приложение", fmt.Sprintf("https://%s", b.config.ServerIP)),
		),
	)

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "HTML"

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

// sendErrorMessage отправляет сообщение об ошибке
func (b *Bot) sendErrorMessage(chatID int64) {
	// Отправляем сообщение об ошибке
	msg := tgbotapi.NewMessage(chatID, "Произошла ошибка. Пожалуйста, попробуйте позже.")
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
