# Документация фронтенда

## Последние обновления

### Исправления от 2024-12-19 (Часть 3)
- ✅ Исправлен delete для боксов - ID теперь передается в body запроса
- ✅ Добавлены реальные переходы по полезным ссылкам между страницами
- ✅ Исправлен API вызов для получения деталей пользователя - используется правильная ручка
- ✅ Добавлена поддержка фильтров по пользователю и номеру бокса в SessionManagement

### Исправления от 2024-12-19 (Часть 2)
- ✅ Исправлены все API вызовы - убраны переменные из путей
- ✅ Все ID теперь передаются в query параметрах или body
- ✅ Соответствие правилам бэкенда (нет переменных в путях)

### Исправления от 2024-12-19 (Часть 1)
- ✅ Исправлена проблема с редактированием номера бокса - теперь номер нельзя изменить при редактировании
- ✅ Добавлена колонка ID в таблицу боксов для лучшей идентификации
- ✅ Исправлены некликабельные ссылки в модальных окнах - добавлены стили и z-index
- ✅ Добавлен реальный API вызов для получения деталей пользователя (getUserById)
- ✅ Все ссылки переходов теперь работают корректно

### Предыдущие исправления
- ✅ Централизованное преобразование camelCase в snake_case во всех API запросах
- ✅ Исправлены фильтры API для корректной работы с бэкендом
- ✅ Добавлены модальные окна для деталей пользователей и сессий
- ✅ Исправлены JSON-схемы при создании боксов
- ✅ Добавлена навигация между сессиями, пользователями и боксами

## Архитектура

### Структура приложения
```
src/
├── apps/
│   ├── admin/          # Админский интерфейс
│   ├── cashier/        # Интерфейс кассира
│   ├── home/           # Главная страница
│   └── telegram/       # Telegram Mini App
├── shared/
│   ├── components/     # Общие компоненты
│   ├── services/       # API сервисы
│   ├── utils/          # Утилиты
│   └── styles/         # Стили и темы
```

### Ключевые компоненты

#### ApiService
Централизованный сервис для работы с API. Автоматически преобразует camelCase в snake_case.

**Важные методы:**
- `getWashBoxes(filters)` - получение списка боксов
- `createWashBox(data)` - создание бокса
- `updateWashBox(id, data)` - обновление бокса (ID в body)
- `deleteWashBox(id)` - удаление бокса (ID в body)
- `getUsers(filters)` - получение списка пользователей
- `getUserById(userId)` - получение деталей пользователя (ID в query)
- `updateUser(id, data)` - обновление пользователя (ID в body)
- `deleteUser(id)` - удаление пользователя (ID в query)
- `getSessions(filters)` - получение списка сессий
- `updateSession(id, data)` - обновление сессии (ID в body)
- `deleteSession(id)` - удаление сессии (ID в query)

#### Утилиты
- `toSnakeCase(obj)` - рекурсивное преобразование ключей в snake_case
- `toSnakeCaseQuery(obj)` - преобразование для query параметров

## Стили и темы

### Тема по умолчанию
```javascript
const lightTheme = {
  primaryColor: '#007bff',
  primaryColorDark: '#0056b3',
  textColor: '#333',
  cardBackground: '#fff',
  // ...
};
```

### Стилизованные компоненты
Все компоненты используют styled-components для консистентного дизайна.

## API Интеграция

### Правила API вызовов
**ВАЖНО**: Все API вызовы соответствуют правилам бэкенда:
- ❌ НЕ используем переменные в путях: `/admin/users/${id}`
- ✅ Используем query параметры: `/admin/users?id=${id}`
- ✅ Используем body для обновлений: `{ ...data, id }`
- ✅ Используем body для DELETE: `{ data: { id } }`

### Примеры правильных вызовов:

```javascript
// ✅ Правильно - ID в query
ApiService.getUserById(userId);
// => GET /admin/users/by-id?id=123

// ✅ Правильно - ID в body
ApiService.updateUser(id, data);
// => PUT /admin/users
// Body: { ...data, id: 123 }

// ✅ Правильно - ID в body для DELETE
ApiService.deleteWashBox(id);
// => DELETE /admin/washboxes
// Body: { id: 123 }

// ❌ Неправильно - ID в пути
// GET /admin/users/123
```

### Автоматическое преобразование форматов
Все API запросы автоматически преобразуют camelCase в snake_case:

```javascript
// Фронтенд (camelCase)
ApiService.createWashBox({
  boxNumber: 1,
  serviceType: 'wash'
});

// Бэкенд получает (snake_case)
{
  "box_number": 1,
  "service_type": "wash"
}
```

### Обработка ошибок
Все API методы включают обработку ошибок и логирование.

## Навигация между страницами

### Полезные ссылки
Реализована система переходов между связанными сущностями:

```javascript
// Переход от пользователя к его сессиям
navigate('/admin/sessions', { 
  state: { 
    filters: { userId: userId },
    showUserFilter: true 
  } 
});

// Переход от бокса к его сессиям
navigate('/admin/sessions', { 
  state: { 
    filters: { boxNumber: boxNumber },
    showBoxFilter: true 
  } 
});

// Переход от сессии к пользователю
navigate(`/users/${sessionDetails.user_id}`);

// Переход от сессии к боксу
navigate(`/boxes/${sessionDetails.box_number}`);
```

### Фильтры в SessionManagement
- Фильтр по ID пользователя
- Фильтр по номеру бокса
- Фильтр по статусу
- Фильтр по типу услуги
- Фильтр по датам

## Компоненты админки

### WashBoxManagement
- Управление боксами мойки
- Создание, редактирование, удаление
- Фильтрация по статусу и типу услуги
- **Номер бокса нельзя редактировать**
- **Все API вызовы соответствуют правилам бэкенда**
- **Реальные переходы к сессиям бокса**

### UserManagement
- Управление пользователями
- Просмотр деталей в модальном окне
- Навигация к сессиям пользователя
- **Реальные API вызовы для получения деталей**
- **ID передается в query параметрах**
- **Реальные переходы к сессиям пользователя**

### SessionManagement
- Управление сессиями
- Детальная информация о сессиях
- Навигация между связанными сущностями
- **Все операции используют правильные API вызовы**
- **Поддержка фильтров по пользователю и боксу**
- **Реальные переходы к пользователям и боксам**

## Безопасность

### Аутентификация
- JWT токены в localStorage
- Автоматическое добавление токена к запросам
- Защищенные маршруты

### Валидация
- Валидация форм на клиенте
- Проверка типов данных
- Санитизация входных данных

## Производительность

### Оптимизации
- Ленивая загрузка компонентов
- Мемоизация дорогих вычислений
- Оптимизированные запросы к API

### Кэширование
- Кэширование данных пользователей
- Локальное хранение настроек
- Оптимизация повторных запросов
