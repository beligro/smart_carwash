.PHONY: build run stop clean

# Переменные
DOCKER_COMPOSE = docker compose

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
	@echo "  make migrate    - Запуск миграций"
	@echo "  make migrate-down - Откат миграций"
