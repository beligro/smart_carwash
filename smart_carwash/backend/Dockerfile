# Используем официальный образ Golang
FROM golang:1.22-alpine AS builder

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

# Копируем бинарный файл из предыдущего этапа
COPY --from=builder /app/main .

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
