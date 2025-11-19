/**
 * ApiService - централизованный сервис для работы с API
 * 
 * ВАЖНО: Все параметры и тела запросов автоматически преобразуются в snake_case
 * перед отправкой на бэкенд. Используйте camelCase в коде фронта!
 * 
 * Примеры:
 * - createWashBox({ boxNumber: 1, serviceType: 'wash' })
 *   => на бэкенд: { box_number: 1, service_type: 'wash' }
 * 
 * - getSessions({ dateFrom: '2024-01-01', serviceType: 'wash' })
 *   => query: ?date_from=2024-01-01&service_type=wash
 */

import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';
import { toSnakeCase, toSnakeCaseQuery } from '../utils/snakeCase';

// Создаем экземпляр axios с базовой конфигурацией
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Добавляем перехватчик для добавления токена к запросам
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Сервис для работы с API
const ApiService = {
  // === МЕТОДЫ ДЛЯ РАБОТЫ С БОКСАМИ ===
  
  // Получение списка боксов
  getWashBoxes: async (filters = {}) => {
    const queryString = toSnakeCaseQuery(filters);
    const response = await api.get(`/admin/washboxes?${queryString}`);
    return response.data;
  },

  // Создание бокса
  createWashBox: async (data) => {
    const snakeData = toSnakeCase(data);
    const response = await api.post('/admin/washboxes', snakeData);
    return response.data;
  },

  // Обновление бокса
  updateWashBox: async (id, data) => {
    const snakeData = toSnakeCase({ ...data, id });
    const response = await api.put('/admin/washboxes', snakeData);
    return response.data;
  },

  // Удаление бокса
  deleteWashBox: async (id) => {
    try {
      const response = await api.delete('/admin/washboxes', { data: { id } });
      return response.data;
    } catch (error) {
      console.error('Ошибка при удалении бокса:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С СЕССИЯМИ ===
  
  // Получение списка сессий
  getSessions: async (filters = {}) => {
    const queryString = toSnakeCaseQuery(filters);
    const response = await api.get(`/admin/sessions?${queryString}`);
    return response.data;
  },

  // Обновление сессии
  updateSession: async (id, data) => {
    const snakeData = toSnakeCase({ ...data, id });
    const response = await api.put('/admin/sessions', snakeData);
    return response.data;
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С ОЧЕРЕДЬЮ ===
  
  // Получение статуса очереди
  getQueueStatus: async () => {
    try {
      const response = await api.get('/queue-status');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении статуса очереди:', error);
      throw error;
    }
  },
  
  // Получение статуса очереди
  getAdminQueue: async () => {
    try {
      const response = await api.get('/admin/queue/status');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении статуса очереди:', error);
      throw error;
    }
  },

  // Получение статуса очереди для кассира
  getCashierQueueStatus: async () => {
    try {
      const response = await api.get('/cashier/queue/status');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении статуса очереди кассира:', error);
      throw error;
    }
  },

  // Удаление из очереди
  removeFromQueue: async (id) => {
    try {
      const response = await api.delete(`/admin/queue?id=${id}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при удалении из очереди:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С ПОЛЬЗОВАТЕЛЯМИ ===
  
  // Получение списка пользователей
  getUsers: async (filters = {}) => {
    const queryString = toSnakeCaseQuery(filters);
    const response = await api.get(`/admin/users?${queryString}`);
    return response.data;
  },

  // Получение пользователя по ID
  getUserById: async (userId) => {
    try {
      const response = await api.get(`/admin/users/by-id?id=${userId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении пользователя по ID:', error);
      throw error;
    }
  },

  // Обновление пользователя
  updateUser: async (id, data) => {
    const snakeData = toSnakeCase({ ...data, id });
    const response = await api.put('/admin/users', snakeData);
    return response.data;
  },

  // Обновление номера машины пользователя
  updateCarNumber: async (userId, carNumber, carNumberCountry = 'RUS') => {
    try {
      const response = await api.put('/users/car-number', {
        user_id: userId,
        car_number: carNumber,
        car_number_country: carNumberCountry
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при обновлении номера машины:', error);
      throw error;
    }
  },


  // === МЕТОДЫ ДЛЯ TELEGRAM ПРИЛОЖЕНИЯ ===

  // Получение доступного времени мойки для определенного типа услуги
  getAvailableRentalTimes: async (serviceType) => {
    try {
      const response = await api.get(`/settings/rental-times?service_type=${serviceType}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении доступного времени мойки:', error);
      return { availableTimes: [15, 20, 30, 60, 90] }; // Значение по умолчанию
    }
  },

  // Получение доступного времени химии для определенного типа услуги (публичный)
  getAvailableChemistryTimes: async (serviceType) => {
    try {
      const response = await api.get(`/settings/available-chemistry-times?service_type=${serviceType}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении доступного времени химии:', error);
      return { available_chemistry_times: [3, 4, 10] }; // Значение по умолчанию
    }
  },

  // Получение пользователя по telegram_id
  getUserByTelegramId: async (telegramId) => {
    try {
      const response = await api.get(`/users/by-telegram-id?telegram_id=${telegramId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении пользователя по telegram_id:', error);
      throw error;
    }
  },

  // Создание пользователя (для Telegram приложения)
  createUser: async (data) => {
    try {
      // Добавляем обязательное поле idempotency_key
      const userData = {
        ...data,
        idempotencyKey: uuidv4() // Генерируем уникальный ключ
      };
      
      const snakeData = toSnakeCase(userData);
      const response = await api.post('/users', snakeData);
      return response.data;
    } catch (error) {
      console.error('Ошибка при создании пользователя:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С ПЛАТЕЖАМИ ===
  
  // Расчет цены услуги
  calculatePrice: async (data) => {
    try {
      const snakeData = toSnakeCase(data);
      const response = await api.post('/payments/calculate-price', snakeData);
      return response.data;
    } catch (error) {
      console.error('Ошибка при расчете цены:', error);
      throw error;
    }
  },

  // Расчет цены продления
  calculateExtensionPrice: async (data) => {
    try {
      const snakeData = toSnakeCase(data);
      const response = await api.post('/payments/calculate-extension-price', snakeData);
      return response.data;
    } catch (error) {
      console.error('Ошибка при расчете цены продления:', error);
      throw error;
    }
  },

  // Создание сессии с платежом
  createSessionWithPayment: async (data) => {
    try {
      // Добавляем обязательное поле idempotency_key
      const sessionData = {
        ...data,
        idempotencyKey: uuidv4() // Генерируем уникальный ключ
      };
      
      const snakeData = toSnakeCase(sessionData);
      const response = await api.post('/sessions/with-payment', snakeData);
      return response.data;
    } catch (error) {
      console.error('Ошибка при создании сессии с платежом:', error);
      
      // Извлекаем сообщение об ошибке из ответа бэкенда
      if (error.response && error.response.data && error.response.data.error) {
        const backendError = new Error(error.response.data.error);
        throw backendError;
      }
      
      throw error;
    }
  },


  // Получение статуса платежа
  getPaymentStatus: async (paymentId) => {
    try {
      const response = await api.get(`/payments/status?payment_id=${paymentId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении статуса платежа:', error);
      throw error;
    }
  },

  // Получение сессии пользователя
  getUserSession: async (userId) => {
    try {
      const response = await api.get(`/sessions?user_id=${userId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении сессии пользователя:', error);
      return { session: null };
    }
  },

  // Получение сессии пользователя для PaymentPage (включая payment_failed)
  getUserSessionForPayment: async (userId) => {
    try {
      const response = await api.get(`/sessions/for-payment?user_id=${userId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении сессии пользователя для PaymentPage:', error);
      return { session: null };
    }
  },

  // Проверка активной сессии пользователя
  checkActiveSession: async (userId) => {
    try {
      const response = await api.get(`/sessions/check-active?user_id=${userId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при проверке активной сессии:', error);
      return { hasActiveSession: false, session: null };
    }
  },
  
  // Получение сессии по ID
  getSessionById: async (sessionId) => {
    try {
      const response = await api.get(`/sessions/by-id?session_id=${sessionId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении сессии по ID:', error);
      return { session: null };
    }
  },
  
  // Запуск сессии (перевод в статус active)
  startSession: async (sessionId) => {
    try {
      const response = await api.post('/sessions/start', { session_id: sessionId });
      return response.data;
    } catch (error) {
      console.error('Ошибка при запуске сессии:', error);
      throw error;
    }
  },
  
  // Завершение сессии (перевод в статус complete)
  completeSession: async (sessionId) => {
    try {
      const response = await api.post('/sessions/complete', { session_id: sessionId });
      return response.data;
    } catch (error) {
      console.error('Ошибка при завершении сессии:', error);
      throw error;
    }
  },

  // Продление сессии с оплатой
  extendSessionWithPayment: async (sessionId, extensionTimeMinutes, extensionChemistryTimeMinutes = 0) => {
    try {
      const requestData = { 
        session_id: sessionId,
        extension_time_minutes: extensionTimeMinutes
      };
      
      // Добавляем время химии только если оно больше 0
      if (extensionChemistryTimeMinutes > 0) {
        requestData.extension_chemistry_time_minutes = extensionChemistryTimeMinutes;
      }
      
      const response = await api.post('/sessions/extend-with-payment', requestData);
      return response.data;
    } catch (error) {
      console.error('Ошибка при продлении сессии с оплатой:', error);
      throw error;
    }
  },

  // Получение платежей сессии
  getSessionPayments: async (sessionId) => {
    try {
      const response = await api.get(`/sessions/payments?session_id=${sessionId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении платежей сессии:', error);
      throw error;
    }
  },

  // Создание нового платежа для существующей сессии
  createNewPayment: async (sessionId, amount, currency) => {
    try {
      const response = await api.post('/payments/create', {
        session_id: sessionId,
        amount: amount,
        currency: currency
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при создании нового платежа:', error);
      throw error;
    }
  },
  
  // Получение истории сессий пользователя
  getUserSessionHistory: async (userId, limit = 5, offset = 0) => {
    try {
      const response = await api.get(`/sessions/history?user_id=${userId}&limit=${limit}&offset=${offset}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении истории сессий:', error);
      throw error;
    }
  },

  // Получить сессию по ID
  async getSession(sessionId) {
    try {
      const response = await api.get(`/sessions?session_id=${sessionId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка получения сессии:', error);
      throw error;
    }
  },

  // Отменить сессию
  async cancelSession(sessionId, userId) {
    try {
      const response = await api.post('/sessions/cancel', {
        session_id: sessionId,
        user_id: userId
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка отмены сессии:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С ПЛАТЕЖАМИ ===
  
  // Получение списка платежей (админка)
  getPayments: async (filters = {}) => {
    const queryString = toSnakeCaseQuery(filters);
    const response = await api.get(`/admin/payments?${queryString}`);
    return response.data;
  },

  // Получение статистики платежей (админка)
  getPaymentStatistics: async (filters = {}) => {
    const queryString = toSnakeCaseQuery(filters);
    const response = await api.get(`/admin/payments/statistics?${queryString}`);
    return response.data;
  },

  // Возврат платежа (админка)
  refundPayment: async (data) => {
    const snakeData = toSnakeCase(data);
    const response = await api.post('/admin/payments/refund', snakeData);
    return response.data;
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С НАСТРОЙКАМИ ===
  
  // Получение настроек сервиса (админка)
  getSettings: async (serviceType) => {
    const response = await api.get(`/admin/settings?service_type=${serviceType}`);
    return response.data;
  },

  // Обновление цен (админка)
  updatePrices: async (data) => {
    const snakeData = toSnakeCase(data);
    const response = await api.put('/admin/settings/prices', snakeData);
    return response.data;
  },

  // Обновление времени мойки (админка)
  updateRentalTimes: async (data) => {
    const snakeData = toSnakeCase(data);
    const response = await api.put('/admin/settings/rental-times', snakeData);
    return response.data;
  },

  // Получение доступного времени химии (админка)
  getAdminAvailableChemistryTimes: async (serviceType) => {
    const response = await api.get(`/admin/settings/available-chemistry-times?service_type=${serviceType}`);
    return response.data;
  },

  // Обновление доступного времени химии (админка)
  updateAvailableChemistryTimes: async (data) => {
    const snakeData = toSnakeCase(data);
    const response = await api.put('/admin/settings/available-chemistry-times', snakeData);
    return response.data;
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С КАССИРОМ ===

  // Получить статус смены кассира
  getShiftStatus: async () => {
    try {
      const response = await api.get('/auth/cashier/shift/status');
      return response.data;
    } catch (error) {
      console.error('Ошибка получения статуса смены:', error);
      throw error;
    }
  },

  // Начать смену кассира
  startShift: async () => {
    try {
      const response = await api.post('/auth/cashier/shift/start');
      return response.data;
    } catch (error) {
      console.error('Ошибка начала смены:', error);
      throw error;
    }
  },

  // Завершить смену кассира
  endShift: async () => {
    try {
      const response = await api.post('/auth/cashier/shift/end');
      return response.data;
    } catch (error) {
      console.error('Ошибка завершения смены:', error);
      throw error;
    }
  },

  // Получить платежи кассира
  getCashierPayments: async (shiftStartedAt, limit = 50, offset = 0) => {
    try {
      const response = await api.get(`/cashier/payments?shift_started_at=${shiftStartedAt}&limit=${limit}&offset=${offset}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка получения платежей кассира:', error);
      throw error;
    }
  },

  // Получить активные сессии кассира
  getCashierActiveSessions: async (limit = 50, offset = 0) => {
    try {
      const response = await api.get(`/cashier/sessions/active?limit=${limit}&offset=${offset}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка получения активных сессий кассира:', error);
      throw error;
    }
  },

  // Запустить сессию кассиром
  startCashierSession: async (sessionId) => {
    try {
      const response = await api.post('/cashier/sessions/start', { session_id: sessionId });
      return response.data;
    } catch (error) {
      console.error('Ошибка запуска сессии кассиром:', error);
      throw error;
    }
  },

  // Завершить сессию кассиром
  completeCashierSession: async (sessionId) => {
    try {
      const response = await api.post('/cashier/sessions/complete', { session_id: sessionId });
      return response.data;
    } catch (error) {
      console.error('Ошибка завершения сессии кассиром:', error);
      throw error;
    }
  },

  // Отменить сессию кассиром
  cancelCashierSession: async (sessionId) => {
    try {
      const response = await api.post('/cashier/sessions/cancel', { 
        session_id: sessionId,
        skip_refund: true 
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка отмены сессии кассиром:', error);
      throw error;
    }
  },

  // Получить статистику последней смены кассира
  getCashierLastShiftStatistics: async () => {
    try {
      const response = await api.get('/cashier/statistics/last-shift');
      return response.data;
    } catch (error) {
      console.error('Ошибка получения статистики последней смены:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С ХИМИЕЙ ===

  // Включить химию в сессии
  enableChemistry: async (sessionId, userId) => {
    try {
      const response = await api.post('/sessions/enable-chemistry', { 
        session_id: sessionId,
        user_id: userId
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка включения химии:', error);
      throw error;
    }
  },

  // Включить химию кассиром
  enableChemistryCashier: async (sessionId) => {
    try {
      const response = await api.post('/cashier/sessions/enable-chemistry', { session_id: sessionId });
      return response.data;
    } catch (error) {
      console.error('Ошибка включения химии кассиром:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С БОКСАМИ КАССИРА ===

  // Получить список боксов для кассира
  getCashierWashBoxes: async (filters = {}) => {
    try {
      const queryString = toSnakeCaseQuery(filters);
      const response = await api.get(`/cashier/washboxes?${queryString}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка получения боксов для кассира:', error);
      throw error;
    }
  },

  // Перевести бокс в режим обслуживания
  setCashierMaintenance: async (boxId) => {
    try {
      const response = await api.post('/cashier/washboxes/maintenance', { id: boxId });
      return response.data;
    } catch (error) {
      console.error('Ошибка перевода бокса на сервис:', error);
      throw error;
    }
  },




  // === МЕТОДЫ ДЛЯ РАБОТЫ С MODBUS ===

  // Тестирование соединения с Modbus устройством
  testModbusConnection: async (boxId) => {
    try {
      const response = await api.post('/admin/modbus/test-connection', { box_id: boxId });
      return response.data;
    } catch (error) {
      console.error('Ошибка тестирования Modbus соединения:', error);
      throw error;
    }
  },

  // Тестирование записи в Modbus coil
  testModbusCoil: async (boxId, register, value) => {
    try {
      const response = await api.post('/admin/modbus/test-coil', { 
        box_id: boxId,
        register: register,
        value: value
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка тестирования Modbus coil:', error);
      throw error;
    }
  },


  // === МЕТОДЫ ДЛЯ ИСТОРИИ ИЗМЕНЕНИЙ БОКСОВ (АДМИНКА) ===
  getWashboxChangeLogs: async (filters = {}) => {
    const queryString = toSnakeCaseQuery(filters);
    const response = await api.get(`/admin/washbox-change-logs?${queryString}`);
    return response.data;
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С УБОРКОЙ ===
  
  // Состояние спецбокса уборщика
  getCleanerBoxState: async () => {
    try {
      const response = await api.get('/cleaner/washboxes/box-state');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении состояния бокса уборщика:', error);
      throw error;
    }
  },
  
  // История уборок конкретного уборщика
  getCleanerCleaningHistory: async (params = {}) => {
    try {
      const queryString = toSnakeCaseQuery(params);
      const response = await api.get(`/cleaner/washboxes/cleaning-history?${queryString}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении истории уборщика:', error);
      throw error;
    }
  },

  // Получение списка боксов для уборщика
  getCleanerWashBoxes: async () => {
    try {
      const response = await api.get('/cleaner/washboxes');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении списка боксов для уборщика:', error);
      throw error;
    }
  },

  // Начало уборки
  startCleaning: async (washBoxId) => {
    try {
      const response = await api.post('/cleaner/washboxes/start-cleaning', {
        wash_box_id: washBoxId,
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при начале уборки:', error);
      throw error;
    }
  },

  // Завершение уборки
  completeCleaning: async (washBoxId) => {
    try {
      const response = await api.post('/cleaner/washboxes/complete-cleaning', {
        wash_box_id: washBoxId,
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при завершении уборки:', error);
      throw error;
    }
  },

  // Получение логов уборки (админка)
  getCleaningLogs: async (filters = {}) => {
    try {
      const response = await api.post('/admin/cleaning-logs', toSnakeCase(filters));
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении логов уборки:', error);
      throw error;
    }
  },

  // Получение времени уборки (админка)
  getCleaningTimeout: async () => {
    try {
      const response = await api.get('/admin/settings/cleaning-timeout');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении времени уборки:', error);
      throw error;
    }
  },

  // Обновление времени уборки (админка)
  updateCleaningTimeout: async (timeoutMinutes) => {
    try {
      const response = await api.put('/admin/settings/cleaning-timeout', {
        timeout_minutes: timeoutMinutes,
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при обновлении времени уборки:', error);
      throw error;
    }
  },

  // Получение времени ожидания старта мойки (админка)
  getSessionTimeout: async () => {
    try {
      const response = await api.get('/admin/settings/session-timeout');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении времени ожидания старта мойки:', error);
      throw error;
    }
  },

  // Обновление времени ожидания старта мойки (админка)
  updateSessionTimeout: async (timeoutMinutes) => {
    try {
      const response = await api.put('/admin/settings/session-timeout', {
        timeout_minutes: timeoutMinutes,
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при обновлении времени ожидания старта мойки:', error);
      throw error;
    }
  },

  // Получение времени блокировки бокса после завершения сессии (админка)
  getCooldownTimeout: async () => {
    try {
      const response = await api.get('/admin/settings/cooldown-timeout');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении времени блокировки бокса:', error);
      throw error;
    }
  },

  // Обновление времени блокировки бокса после завершения сессии (админка)
  updateCooldownTimeout: async (timeoutMinutes) => {
    try {
      const response = await api.put('/admin/settings/cooldown-timeout', {
        timeout_minutes: timeoutMinutes,
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при обновлении времени блокировки бокса:', error);
      throw error;
    }
  },

  // Переназначение сессии на другой бокс (админка)
  adminReassignSession: async (sessionId) => {
    try {
      const response = await api.post('/admin/sessions/reassign', {
        session_id: sessionId,
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при переназначении сессии администратором:', error);
      throw error;
    }
  },

  // Отменить сессию администратором
  adminCancelSession: async (sessionId) => {
    try {
      const response = await api.post('/admin/sessions/cancel', { 
        session_id: sessionId
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка отмены сессии администратором:', error);
      throw error;
    }
  },

  // Переназначение сессии на другой бокс (кассир)
  cashierReassignSession: async (sessionId) => {
    try {
      const response = await api.post('/cashier/sessions/reassign', {
        session_id: sessionId,
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при переназначении сессии кассиром:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ СО СТАТУСОМ МОЙКИ ===
  
  // Получение текущего статуса мойки (публичный)
  getCarwashStatus: async () => {
    const response = await api.get('/carwash/status');
    return response.data;
  },

  // Получение текущего статуса мойки (админка)
  getCarwashStatusAdmin: async () => {
    const response = await api.get('/admin/carwash/status');
    return response.data;
  },

  // Закрытие мойки (админка)
  closeCarwash: async (reason) => {
    const response = await api.post('/admin/carwash/close', { reason: reason || null });
    return response.data;
  },

  // Открытие мойки (админка)
  openCarwash: async () => {
    const response = await api.post('/admin/carwash/open', {});
    return response.data;
  },

  // Получение истории изменений статуса мойки (админка)
  getCarwashStatusHistory: async (filters = {}) => {
    const queryString = toSnakeCaseQuery(filters);
    const response = await api.get(`/admin/carwash/history?${queryString}`);
    return response.data;
  },

};

// Добавляем перехватчик для обработки ошибок
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Обрабатываем ошибки
    if (error.response) {
      // Ошибка от сервера
      // Если сервер вернул 401, значит токен истек или недействителен
      if (error.response.status === 401) {
        // Удаляем токен и перенаправляем на страницу входа
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('expiresAt');
        
        // Определяем, какой тип пользователя был авторизован
        const isAdmin = localStorage.getItem('isAdmin') === 'true';
        localStorage.removeItem('isAdmin');
        
        // Перенаправляем на соответствующую страницу входа
        if (isAdmin) {
          window.location.href = '/admin/login';
        } else {
          window.location.href = '/cashier/login';
        }
        
        // Не продолжаем выполнение - запрос завершается здесь
        return Promise.reject(error);
      }
      
      console.error('Ошибка сервера:', error.response.data);
    } else if (error.request) {
      // Нет ответа от сервера
      console.error('Нет ответа от сервера:', error.request);
    } else {
      // Ошибка при настройке запроса
      console.error('Ошибка запроса:', error.message);
    }
    
    return Promise.reject(error);
  }
);

export default ApiService; 