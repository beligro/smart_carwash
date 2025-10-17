package logger

// Пакет logger предоставляет структурированное логирование с поддержкой trace_id
//
// Использование:
//
// 1. В handlers (с gin.Context):
//    logger.WithContext(c).Info("Processing request")
//    logger.WithContext(c).Errorf("Error: %v", err)
//
// 2. В service/repository (без context):
//    logger.Info("Starting operation")
//    logger.WithFields(logrus.Fields{"user_id": userID}).Info("User found")
//
// 3. С явным trace_id:
//    logger.WithTraceID(traceID).Info("Background job processing")
//
// Middleware автоматически добавляет trace_id в каждый запрос.

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range hook.Writer {
		_, err = w.Write([]byte(line))
	}
	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

var e *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func GetLogger() Logger {
	return Logger{e}
}

func (l *Logger) GetLoggerWithField(k string, v interface{}) Logger {
	return Logger{l.WithField(k, v)}
}

func Init() {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
		},
	}

	// Создаем директорию для логов
	err := os.MkdirAll("/var/log/backend", 0o755)
	if err != nil && !os.IsExist(err) {
		panic("can't create log dir. no configured logging to files")
	}

	// Открываем файл для записи логов
	allFile, err := os.OpenFile("/var/log/backend/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o660)
	if err != nil {
		panic(fmt.Sprintf("[Error]: %s", err))
	}

	l.SetOutput(io.Discard) // Отправляем все логи в никуда по умолчанию

	// Добавляем hook для записи и в файл, и в stdout
	l.AddHook(&writerHook{
		Writer:    []io.Writer{allFile, os.Stdout},
		LogLevels: logrus.AllLevels,
	})

	l.SetLevel(logrus.InfoLevel)

	e = logrus.NewEntry(l)
}

// WithFields создает логгер с дополнительными полями
func WithFields(fields logrus.Fields) *logrus.Entry {
	return e.WithFields(fields)
}

// Info логирует информационное сообщение
func Info(args ...interface{}) {
	e.Info(args...)
}

// Error логирует ошибку
func Error(args ...interface{}) {
	e.Error(args...)
}

// Warn логирует предупреждение
func Warn(args ...interface{}) {
	e.Warn(args...)
}

// Debug логирует отладочное сообщение
func Debug(args ...interface{}) {
	e.Debug(args...)
}

// Fatal логирует критическую ошибку и завершает программу
func Fatal(args ...interface{}) {
	e.Fatal(args...)
}

// Printf логирует сообщение в формате Printf (для совместимости)
func Printf(format string, args ...interface{}) {
	e.Infof(format, args...)
}

// WithContext создает логгер с trace_id из gin.Context
func WithContext(c *gin.Context) *logrus.Entry {
	if c == nil {
		return e
	}
	
	traceID, exists := c.Get("trace_id")
	if !exists {
		return e
	}
	
	return e.WithField("trace_id", traceID)
}

// WithTraceID создает логгер с указанным trace_id
func WithTraceID(traceID string) *logrus.Entry {
	return e.WithField("trace_id", traceID)
}
