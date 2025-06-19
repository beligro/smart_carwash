import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';

// Создаем экземпляр axios с базовым URL
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});


// Сервис для работы с API
const ApiService = {
  // Получение доступного времени аренды для определенного типа услуги
  getAvailableRentalTimes: async (serviceType) => {
    try {
      const response = await api.get(`/settings/rental-times?service_type=${serviceType}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении доступного времени аренды:', error);
      return { available_times: [5] }; // Значение по умолчанию
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

  // Получение статуса очереди и боксов
  getQueueStatus: async () => {
    try {
      const response = await api.get('/queue-status');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении статуса очереди:', error);
      throw error;
    }
  },

  // Создание пользователя
  createUser: async (userData) => {
    try {
      // Добавляем токен идемпотентности
      const dataWithToken = {
        ...userData,
        idempotency_key: uuidv4()
      };
      
      const response = await api.post('/users', dataWithToken);
      return response.data;
    } catch (error) {
      console.error('Ошибка при создании пользователя:', error);
      throw error;
    }
  },

  // Создание сессии
  createSession: async (sessionData) => {
    try {
      // Добавляем токен идемпотентности
      const dataWithToken = {
        ...sessionData,
        idempotency_key: uuidv4()
      };
      
      const response = await api.post('/sessions', dataWithToken);
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

export default ApiService;
