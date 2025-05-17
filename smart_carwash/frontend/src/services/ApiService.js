import axios from 'axios';

// Создаем экземпляр axios с базовым URL
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});


// Сервис для работы с API
const ApiService = {
  // Получение информации о мойке
  getWashInfo: async () => {
    try {
      const response = await api.get('/wash-info');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении информации о мойке:', error);
      throw error;
    }
  },

  // Получение информации о мойке для конкретного пользователя
  getWashInfoForUser: async (userId) => {
    try {
      const response = await api.get(`/wash-info/${userId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении информации о мойке для пользователя:', error);
      throw error;
    }
  },

  // Создание пользователя
  createUser: async (userData) => {
    try {
      const response = await api.post('/users', userData);
      return response.data;
    } catch (error) {
      console.error('Ошибка при создании пользователя:', error);
      throw error;
    }
  },

  // Создание сессии
  createSession: async (sessionData) => {
    try {
      const response = await api.post('/sessions', sessionData);
      return response.data;
    } catch (error) {
      console.error('Ошибка при создании сессии:', error);
      throw error;
    }
  },

  // Получение сессии пользователя
  getUserSession: async (userId) => {
    try {
      const response = await api.get(`/sessions/${userId}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении сессии пользователя:', error);
      return null;
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
