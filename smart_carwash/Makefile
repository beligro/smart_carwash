.PHONY: build run stop clean

# Переменные
DOCKER_COMPOSE = docker-compose

# Сборка проекта
build:
	$(DOCKER_COMPOSE) build

# Запуск проекта
run:
	$(DOCKER_COMPOSE) up -d

# Остановка проекта
stop:
	$(DOCKER_COMPOSE) down

# Перезапуск проекта
restart: stop run

# Просмотр логов
logs:
	$(DOCKER_COMPOSE) logs -f

# Очистка проекта
clean: stop
	$(DOCKER_COMPOSE) down -v --remove-orphans

# Создание SSL сертификатов для разработки
ssl-dev:
	mkdir -p nginx/ssl
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout nginx/ssl/privkey.pem \
		-out nginx/ssl/fullchain.pem \
		-subj "/C=RU/ST=Moscow/L=Moscow/O=CarWash/CN=158.160.105.190"

# Запуск миграций
migrate:
	$(DOCKER_COMPOSE) exec backend sh -c "cd /app && go run github.com/golang-migrate/migrate/v4/cmd/migrate -path=./migrations -database=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable up"

# Откат миграций
migrate-down:
	$(DOCKER_COMPOSE) exec backend sh -c "cd /app && go run github.com/golang-migrate/migrate/v4/cmd/migrate -path=./migrations -database=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable down"

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make build      - Сборка проекта"
	@echo "  make run        - Запуск проекта"
	@echo "  make stop       - Остановка проекта"
	@echo "  make restart    - Перезапуск проекта"
	@echo "  make logs       - Просмотр логов"
	@echo "  make clean      - Очистка проекта"
	@echo "  make ssl-dev    - Создание SSL сертификатов для разработки"
	@echo "  make migrate    - Запуск миграций"
	@echo "  make migrate-down - Откат миграций"
