# Этап сборки
FROM node:18-alpine AS build

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем package.json и package-lock.json
COPY package*.json ./

# Устанавливаем зависимости
RUN npm install

# Копируем исходный код
COPY . .

# Аргумент для выбора приложения (по умолчанию - все приложения)
ARG APP_TYPE=all

# Собираем приложение в зависимости от типа
RUN if [ "$APP_TYPE" = "telegram" ]; then \
        echo "Building Telegram Mini App only"; \
        REACT_APP_STANDALONE=true REACT_APP_APP_TYPE=telegram npm run build; \
    elif [ "$APP_TYPE" = "admin" ]; then \
        echo "Building Admin App only"; \
        REACT_APP_STANDALONE=true REACT_APP_APP_TYPE=admin npm run build; \
    elif [ "$APP_TYPE" = "cashier" ]; then \
        echo "Building Cashier App only"; \
        REACT_APP_STANDALONE=true REACT_APP_APP_TYPE=cashier npm run build; \
    else \
        echo "Building all apps"; \
        npm run build; \
    fi

# Этап запуска
FROM nginx:alpine

# Копируем собранное приложение из предыдущего этапа
COPY --from=build /app/build /usr/share/nginx/html

# Копируем конфигурацию nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Открываем порт
EXPOSE 80

# Запускаем nginx
CMD ["nginx", "-g", "daemon off;"]
