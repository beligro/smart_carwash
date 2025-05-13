import axios from 'axios';

// Создаем экземпляр axios с базовым URL
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'https://158.160.105.190/api',
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
