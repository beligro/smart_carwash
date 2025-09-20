# 🏭 ПЛК210 Веб-сервер для Артема

**Дата:** 17 сентября 2025  
**Версия:** 1.0  
**Разработчик:** Александр (ginko.press)

## 📋 Описание

Полнофункциональный веб-сервер для управления ПЛК210 через Modbus TCP. 
Предоставляет REST API и веб-интерфейс для управления 90 битами в 6 Holding Registers.

## 🔧 Технические характеристики

### Подключение к ПЛК:
- **IP:** 195.208.131.189
- **Порт:** 502
- **Unit ID:** 1
- **Протокол:** Modbus TCP
- **Таймаут:** 5 секунд

### Управляемые ресурсы:
- **Holding Registers:** 0, 1, 2, 3, 4, 5
- **Всего битов:** 90 (Register 0-4: по 16 битов, Register 5: 10 битов)
- **Тип данных:** 16-битные слова (WORD)

## 🚀 Быстрый старт

### 1. Установка зависимостей:
```bash
pip install -r requirements.txt
```

### 2. Запуск сервера:
```bash
python3 plc210_webserver.py
```

### 3. Открыть в браузере:
```
http://localhost:5000
```

## 📡 REST API Endpoints

### Получение состояния всех регистров
```http
GET /api/status
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "timestamp": "2025-09-17T22:30:00.123456",
    "registers": {
      "register_0": {
        "value": 1234,
        "bits": [0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 1, 0, 0]
      },
      // ... остальные регистры
    },
    "connection_status": true,
    "last_error": null
  }
}
```

### Установка состояния бита
```http
POST /api/set_bit
Content-Type: application/json

{
  "register": 0,
  "bit": 5,
  "state": true
}
```

**Ответ:**
```json
{
  "success": true,
  "message": "Бит 5 в Register 0 установлен в ВКЛ"
}
```

### Получение состояния бита
```http
GET /api/get_bit?register=0&bit=5
```

**Ответ:**
```json
{
  "success": true,
  "register": 0,
  "bit": 5,
  "state": true
}
```

### Чтение значения регистра
```http
GET /api/read_register?register=0
```

**Ответ:**
```json
{
  "success": true,
  "register": 0,
  "value": 1234,
  "hex": "0x04D2",
  "binary": "0000010011010010",
  "bits": [0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 1, 0, 0]
}
```

### Запись значения в регистр
```http
POST /api/write_register
Content-Type: application/json

{
  "register": 0,
  "value": 1234
}
```

### Включение/выключение всех битов
```http
POST /api/turn_all
Content-Type: application/json

{
  "state": true
}
```

**Ответ:**
```json
{
  "success": true,
  "message": "Все 90 битов успешно включены",
  "success_count": 90,
  "total_count": 90
}
```

### Информация о подключении
```http
GET /api/connection
```

**Ответ:**
```json
{
  "success": true,
  "connection": {
    "host": "195.208.131.189",
    "port": 502,
    "unit_id": 1,
    "timeout": 5.0,
    "connected": true,
    "last_error": null
  }
}
```

## 🎯 Примеры использования в коде

### Python (requests):
```python
import requests

# Включить бит 0 в Register 0
response = requests.post('http://localhost:5000/api/set_bit', 
                        json={'register': 0, 'bit': 0, 'state': True})
print(response.json())

# Получить состояние всех регистров
response = requests.get('http://localhost:5000/api/status')
data = response.json()
if data['success']:
    registers = data['data']['registers']
    print(f"Register 0 value: {registers['register_0']['value']}")
```

### JavaScript (fetch):
```javascript
// Выключить бит 5 в Register 2
fetch('/api/set_bit', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({register: 2, bit: 5, state: false})
})
.then(response => response.json())
.then(data => console.log(data));

// Получить состояние конкретного бита
fetch('/api/get_bit?register=1&bit=10')
.then(response => response.json())
.then(data => {
    if (data.success) {
        console.log(`Bit ${data.bit} in Register ${data.register}: ${data.state}`);
    }
});
```

### cURL:
```bash
# Включить все биты
curl -X POST http://localhost:5000/api/turn_all \
     -H "Content-Type: application/json" \
     -d '{"state": true}'

# Прочитать Register 3
curl "http://localhost:5000/api/read_register?register=3"

# Записать значение 0xAAAA в Register 1
curl -X POST http://localhost:5000/api/write_register \
     -H "Content-Type: application/json" \
     -d '{"register": 1, "value": 43690}'
```

## 🗺️ Маппинг битов

### Структура регистров:
- **Register 0:** Биты 0-15 (16 битов)
- **Register 1:** Биты 0-15 (16 битов)
- **Register 2:** Биты 0-15 (16 битов)
- **Register 3:** Биты 0-15 (16 битов)
- **Register 4:** Биты 0-15 (16 битов)
- **Register 5:** Биты 0-9 (10 битов, биты 10-15 не используются)

### Общий индекс битов:
```
Общий бит 0  = Register 0, Bit 0
Общий бит 15 = Register 0, Bit 15
Общий бит 16 = Register 1, Bit 0
...
Общий бит 89 = Register 5, Bit 9
```

## ⚠️ Важные особенности

### 1. **Обратная связь:**
- ❌ **Input Registers недоступны** (не настроены в CODESYS)
- ❌ **Нет чтения текущего состояния** с физических выходов
- ✅ **Можно читать только записанные значения** в Holding Registers

### 2. **Ответы ПЛК:**
- ✅ **Запись работает** (команды выполняются в CODESYS)
- ⚠️ **Ответы могут быть неполными** (это нормально для данного ПЛК)
- ✅ **Изменения видны в CODESYS** в реальном времени

### 3. **Соединение:**
- ✅ **Автоматическое переподключение** при разрывах связи
- ✅ **Thread-safe** операции
- ✅ **Таймауты и обработка ошибок**

## 🔍 Мониторинг и отладка

### Логи:
- **Файл:** `plc210_webserver.log`
- **Формат:** Timestamp - Level - Message
- **Уровни:** INFO, WARNING, ERROR

### Веб-интерфейс:
- **URL:** http://localhost:5000
- **Автообновление** каждые 2 секунды
- **Визуальное управление** всеми 90 битами
- **Статус подключения** в реальном времени

## 🛠️ Конфигурация

### Изменение IP/порта:
```python
# В файле plc210_webserver.py, строка ~90
config = ModbusConfig(
    host="195.208.131.189",  # <- Изменить IP
    port=502,                # <- Изменить порт
    unit_id=1,
    timeout=5.0
)
```

### Изменение веб-порта:
```python
# В функции main(), строка ~850
app.run(
    host='0.0.0.0',
    port=5000,  # <- Изменить веб-порт
    debug=False,
    threaded=True
)
```

## 📞 Поддержка

**Разработчик:** Александр  
**Telegram:** @ginko  
**Email:** ginko.press@gmail.com

## 📝 История изменений

### v1.0 (17.09.2025)
- ✅ Первая рабочая версия
- ✅ REST API для всех операций
- ✅ Веб-интерфейс с автообновлением
- ✅ Управление 90 битами в 6 регистрах
- ✅ Автоматическое переподключение
- ✅ Подробное логирование

---

**🎯 Готово к продакшену!** 🚀
