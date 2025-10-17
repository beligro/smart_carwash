# Используем официальный образ Golang
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Используем минимальный образ для запуска
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Создаем директорию для логов
RUN mkdir -p /var/log/backend

# Копируем бинарный файл из предыдущего этапа
COPY --from=builder /app/main .

# Копируем директорию с миграциями
COPY --from=builder /app/migrations ./migrations

# Открываем порт
EXPOSE 8080

# Запускаем приложение с перенаправлением логов
CMD ["./main"]
