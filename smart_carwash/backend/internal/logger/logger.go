package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Константы для ключей контекста
const (
	TraceIDKey    = "trace_id"
	UserIDKey     = "user_id"
	TelegramIDKey = "telegram_id"
	SessionIDKey  = "session_id"
	BoxIDKey      = "box_id"
)

// Config содержит настройки логгера
type Config struct {
	Level      string
	Pretty     bool
	TimeFormat string
}

var (
	// DefaultConfig - конфигурация логгера по умолчанию
	DefaultConfig = Config{
		Level:      "info",
		Pretty:     false,
		TimeFormat: time.RFC3339,
	}

	// log - глобальный экземпляр логгера
	log zerolog.Logger
)

// Init инициализирует логгер с заданной конфигурацией
func Init(cfg Config) {
	// Настраиваем уровень логирования
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Настраиваем форматирование ошибок
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Настраиваем вывод
	var output io.Writer = os.Stdout
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: cfg.TimeFormat,
		}
	}

	// Создаем логгер
	log = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()
}

// WithTraceID добавляет trace_id в контекст
func WithTraceID(ctx context.Context) context.Context {
	traceID := uuid.New().String()
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithUserID добавляет user_id в контекст
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithTelegramID добавляет telegram_id в контекст
func WithTelegramID(ctx context.Context, telegramID string) context.Context {
	return context.WithValue(ctx, TelegramIDKey, telegramID)
}

// WithSessionID добавляет session_id в контекст
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// WithBoxID добавляет box_id в контекст
func WithBoxID(ctx context.Context, boxID string) context.Context {
	return context.WithValue(ctx, BoxIDKey, boxID)
}

// FromContext создает логгер с данными из контекста
func FromContext(ctx context.Context) *zerolog.Logger {
	logger := log.With()

	// Добавляем trace_id, если есть
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		logger = logger.Str("trace_id", traceID)
	}

	// Добавляем user_id, если есть
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		logger = logger.Str("user_id", userID)
	}

	// Добавляем telegram_id, если есть
	if telegramID, ok := ctx.Value(TelegramIDKey).(string); ok {
		logger = logger.Str("telegram_id", telegramID)
	}

	// Добавляем session_id, если есть
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok {
		logger = logger.Str("session_id", sessionID)
	}

	// Добавляем box_id, если есть
	if boxID, ok := ctx.Value(BoxIDKey).(string); ok {
		logger = logger.Str("box_id", boxID)
	}

	l := logger.Logger()
	return &l
}

// Debug логирует сообщение с уровнем Debug
func Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := FromContext(ctx)
	event := logger.Debug()
	addFields(event, fields...)
	event.Msg(msg)
}

// Info логирует сообщение с уровнем Info
func Info(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := FromContext(ctx)
	event := logger.Info()
	addFields(event, fields...)
	event.Msg(msg)
}

// Warn логирует сообщение с уровнем Warn
func Warn(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := FromContext(ctx)
	event := logger.Warn()
	addFields(event, fields...)
	event.Msg(msg)
}

// Error логирует сообщение с уровнем Error
func Error(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	logger := FromContext(ctx)
	event := logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	addFields(event, fields...)
	event.Msg(msg)
}

// Fatal логирует сообщение с уровнем Fatal
func Fatal(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	logger := FromContext(ctx)
	event := logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	addFields(event, fields...)
	event.Msg(msg)
}

// addFields добавляет дополнительные поля в событие логирования
func addFields(event *zerolog.Event, fields ...map[string]interface{}) {
	for _, field := range fields {
		for k, v := range field {
			event = event.Interface(k, v)
		}
	}
}
