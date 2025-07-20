package telegram

import (
	"fmt"
	"log"
	"strings"

	"carwash_backend/internal/domain/user/models"
	"carwash_backend/internal/domain/user/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotHandler обрабатывает команды бота
type BotHandler struct {
	bot         *tgbotapi.BotAPI
	webAppURL   string
	userService service.Service
}

// NewBotHandler создает новый обработчик бота
func NewBotHandler(bot *tgbotapi.BotAPI, webAppURL string, userService service.Service) *BotHandler {
	return &BotHandler{
		bot:         bot,
		webAppURL:   webAppURL,
		userService: userService,
	}
}

// HandleStartCommand обрабатывает команду /start
func (h *BotHandler) HandleStartCommand(update tgbotapi.Update) {
	message := update.Message
	if message == nil {
		return
	}

	// Получаем параметры из команды /start
	text := message.Text
	if !strings.HasPrefix(text, "/start") {
		return
	}

	// Извлекаем параметры после /start
	parts := strings.Fields(text)
	if len(parts) < 2 {
		// Обычная команда /start без параметров
		h.createUserAndSendWelcome(message)
		return
	}

	startParam := parts[1]
	log.Printf("Получена команда /start с параметром: %s", startParam)

	// Обрабатываем параметры (fallback для startapp)
	switch startParam {
	case "payment_success":
		h.handlePaymentSuccess(message.Chat.ID, message.From.ID)
	case "payment_fail":
		h.handlePaymentFail(message.Chat.ID, message.From.ID)
	default:
		// Создаем пользователя при обычной команде /start
		h.createUserAndSendWelcome(message)
	}
}

// handlePaymentSuccess обрабатывает успешный платеж (fallback)
func (h *BotHandler) handlePaymentSuccess(chatID int64, userID int64) {
	log.Printf("Fallback: обработка успешного платежа для пользователя %d", userID)

	// Создаем кнопку для открытия Mini App
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🎉 Платеж успешен! Открыть приложение",
				"open_mini_app_success",
			),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "✅ Платеж успешно завершен!\n\nВы были добавлены в очередь. Нажмите кнопку ниже, чтобы открыть приложение и посмотреть статус.")
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения об успешном платеже: %v", err)
	}
}

// handlePaymentFail обрабатывает неудачный платеж (fallback)
func (h *BotHandler) handlePaymentFail(chatID int64, userID int64) {
	log.Printf("Fallback: обработка неудачного платежа для пользователя %d", userID)

	// Создаем кнопку для открытия Mini App
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Платеж не завершен. Попробовать снова",
				"open_mini_app_fail",
			),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "❌ Платеж не был завершен.\n\nПопробуйте еще раз или обратитесь в поддержку.")
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения о неудачном платеже: %v", err)
	}
}

// createUserAndSendWelcome создает пользователя и отправляет приветственное сообщение
func (h *BotHandler) createUserAndSendWelcome(message *tgbotapi.Message) {
	// Создаем пользователя
	user, err := h.userService.CreateUser(&models.CreateUserRequest{
		TelegramID: message.From.ID,
		Username:   message.From.UserName,
		FirstName:  message.From.FirstName,
		LastName:   message.From.LastName,
	})

	if err != nil {
		log.Printf("Ошибка создания пользователя: %v", err)
		h.sendErrorMessage(message.Chat.ID)
		return
	}

	// Отправляем приветственное сообщение с кнопкой
	h.sendWelcomeMessage(message.Chat.ID, user.FirstName)
}

// sendWelcomeMessage отправляет приветственное сообщение
func (h *BotHandler) sendWelcomeMessage(chatID int64, firstName string) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🚗 Открыть приложение",
				"open_mini_app",
			),
		),
	)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("👋 Привет, %s!\n\nДобро пожаловать в Smart Carwash!\n\nНажмите кнопку ниже, чтобы открыть приложение и записаться на мойку.", firstName))
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки приветственного сообщения: %v", err)
	}
}

// sendErrorMessage отправляет сообщение об ошибке
func (h *BotHandler) sendErrorMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Произошла ошибка. Пожалуйста, попробуйте позже.")
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения об ошибке: %v", err)
	}
}

// HandleCallbackQuery обрабатывает нажатия на inline кнопки
func (h *BotHandler) HandleCallbackQuery(update tgbotapi.Update) {
	callback := update.CallbackQuery
	if callback == nil {
		return
	}

	log.Printf("Получен callback: %s", callback.Data)

	switch callback.Data {
	case "open_mini_app", "open_mini_app_success", "open_mini_app_fail":
		h.openMiniApp(callback)
	default:
		// Отвечаем на неизвестный callback
		h.bot.Request(tgbotapi.NewCallback(callback.ID, "Неизвестная команда"))
	}
}

// openMiniApp открывает Mini App
func (h *BotHandler) openMiniApp(callback *tgbotapi.CallbackQuery) {
	// Создаем обычную кнопку для открытия Mini App
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🚗 Открыть приложение",
				"open_mini_app",
			),
		),
	)

	// Отправляем новое сообщение с кнопкой
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Нажмите кнопку ниже, чтобы открыть приложение:")
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки кнопки: %v", err)
		h.bot.Request(tgbotapi.NewCallback(callback.ID, "Ошибка открытия приложения"))
		return
	}

	// Отвечаем на callback
	h.bot.Request(tgbotapi.NewCallback(callback.ID, "Открываю приложение..."))
} 