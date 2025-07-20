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
  baseURL: process.env.REACT_APP_API_URL || '/api/v1',
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
  updateCarNumber: async (userId, carNumber) => {
    try {
      const response = await api.put('/users/car-number', {
        user_id: userId,
        car_number: carNumber
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при обновлении номера машины:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ TELEGRAM ПРИЛОЖЕНИЯ ===

  // Получение доступного времени аренды для определенного типа услуги
  getAvailableRentalTimes: async (serviceType) => {
    try {
      const response = await api.get(`/settings/rental-times?service_type=${serviceType}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении доступного времени аренды:', error);
      return { availableTimes: [5] }; // Значение по умолчанию
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

  // Создание сессии (для Telegram приложения)
  createSession: async (data) => {
    try {
      // Добавляем обязательное поле idempotency_key
      const sessionData = {
        ...data,
        idempotencyKey: uuidv4() // Генерируем уникальный ключ
      };
      
      const snakeData = toSnakeCase(sessionData);
      const response = await api.post('/sessions', snakeData);
      return response.data;
    } catch (error) {
      console.error('Ошибка при создании сессии:', error);
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
  
  // Получение платежа по ID
  getPaymentById: async (paymentId) => {
    try {
      const response = await api.get(`/payments/by-id?id=${paymentId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении платежа по ID:', error);
      return { payment: null };
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
  
  // Продление сессии (добавление времени к активной сессии)
  extendSession: async (sessionId, extensionTimeMinutes) => {
    try {
      const response = await api.post('/sessions/extend', { 
        session_id: sessionId,
        extension_time_minutes: extensionTimeMinutes
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при продлении сессии:', error);
      throw error;
    }
  },
  
  // Получение истории сессий пользователя
  getUserSessionHistory: async (userId, limit = 10, offset = 0) => {
    try {
      const response = await api.get(`/sessions/history?user_id=${userId}&limit=${limit}&offset=${offset}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении истории сессий:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С НАСТРОЙКАМИ ===
  
  // Получение настройки
  getSetting: async (serviceType, settingKey) => {
    try {
      const response = await api.get(`/settings/service-setting?service_type=${serviceType}&setting_key=${settingKey}`);
      return response.data.value;
    } catch (error) {
      console.error('Ошибка при получении настройки:', error);
      return null;
    }
  },

  // Обновление настройки
  updateSetting: async (serviceType, settingKey, value) => {
    try {
      const response = await api.put('/settings/service-setting', {
        service_type: serviceType,
        setting_key: settingKey,
        setting_value: value
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при обновлении настройки:', error);
      throw error;
    }
  },

  // Обновление доступного времени аренды
  updateAvailableRentalTimes: async (serviceType, availableTimes) => {
    try {
      const response = await api.put('/settings/rental-times', {
        service_type: serviceType,
        available_times: availableTimes
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка при обновлении времени аренды:', error);
      throw error;
    }
  },

  // Получение всех настроек
  getAllSettings: async () => {
    try {
      const response = await api.get('/settings/all');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении всех настроек:', error);
      throw error;
    }
  },

  // === МЕТОДЫ ДЛЯ РАБОТЫ С ПЛАТЕЖАМИ ===
  
  // Получение списка платежей
  getPayments: async (filters = {}) => {
    try {
      const queryString = toSnakeCaseQuery(filters);
      const url = `/admin/payments?${queryString}`;
      console.log('Отправляем запрос на:', url);
      const response = await api.get(url);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении списка платежей:', error);
      throw error;
    }
  },

  // Получение деталей платежа
  getPaymentDetails: async (paymentId) => {
    try {
      const url = `/admin/payments/by-id?id=${paymentId}`;
      console.log('Отправляем запрос на:', url);
      const response = await api.get(url);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении деталей платежа:', error);
      throw error;
    }
  },
};

// Добавляем перехватчик для обработки ошибок
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Обрабатываем ошибки
    if (error.response) {
      // Ошибка от сервера
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

export { api };
export default ApiService; 