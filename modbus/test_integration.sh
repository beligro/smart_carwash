#!/bin/bash

# Скрипт для тестирования интеграции modbus сервера

BASE_URL="http://localhost:8081"
BOX_ID="123e4567-e89b-12d3-a456-426614174000"

echo "🧪 Тестирование Modbus HTTP Server"
echo "=================================="

# Проверка health check
echo "1. Проверка health check..."
curl -s "$BASE_URL/health" | jq .
echo ""

# Тест соединения
echo "2. Тест соединения с Modbus..."
curl -s -X POST "$BASE_URL/api/v1/modbus/test-connection" \
  -H "Content-Type: application/json" \
  -d "{\"box_id\": \"$BOX_ID\"}" | jq .
echo ""

# Тест управления светом (включить)
echo "3. Включение света..."
curl -s -X POST "$BASE_URL/api/v1/modbus/light" \
  -H "Content-Type: application/json" \
  -d "{\"box_id\": \"$BOX_ID\", \"value\": true}" | jq .
echo ""

# Ждем 2 секунды
sleep 2

# Тест управления светом (выключить)
echo "4. Выключение света..."
curl -s -X POST "$BASE_URL/api/v1/modbus/light" \
  -H "Content-Type: application/json" \
  -d "{\"box_id\": \"$BOX_ID\", \"value\": false}" | jq .
echo ""

# Тест управления химией (включить)
echo "5. Включение химии..."
curl -s -X POST "$BASE_URL/api/v1/modbus/chemistry" \
  -H "Content-Type: application/json" \
  -d "{\"box_id\": \"$BOX_ID\", \"value\": true}" | jq .
echo ""

# Ждем 2 секунды
sleep 2

# Тест управления химией (выключить)
echo "6. Выключение химии..."
curl -s -X POST "$BASE_URL/api/v1/modbus/chemistry" \
  -H "Content-Type: application/json" \
  -d "{\"box_id\": \"$BOX_ID\", \"value\": false}" | jq .
echo ""

# Тест записи в coil
echo "7. Тест записи в coil..."
curl -s -X POST "$BASE_URL/api/v1/modbus/coil" \
  -H "Content-Type: application/json" \
  -d "{\"box_id\": \"$BOX_ID\", \"register\": \"0x001\", \"value\": true}" | jq .
echo ""

# Тест coil
echo "8. Тест coil..."
curl -s -X POST "$BASE_URL/api/v1/modbus/test-coil" \
  -H "Content-Type: application/json" \
  -d "{\"box_id\": \"$BOX_ID\", \"register\": \"0x001\", \"value\": true}" | jq .
echo ""

echo "✅ Тестирование завершено!"
