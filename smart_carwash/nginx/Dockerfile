# Используем официальный образ Nginx
FROM nginx:alpine

# Копируем конфигурацию Nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Создаем директорию для SSL сертификатов
RUN mkdir -p /etc/nginx/ssl

# Открываем порты
EXPOSE 80 443

# Запускаем Nginx
CMD ["nginx", "-g", "daemon off;"]
