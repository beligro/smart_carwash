import axios from 'axios';

// Создаем экземпляр axios с базовым URL
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || '/api',
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

// Добавляем перехватчик для обработки ошибок авторизации
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      // Если сервер вернул 401, значит токен истек или недействителен
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
    }
    return Promise.reject(error);
  }
);

// Сервис для работы с авторизацией
const AuthService = {
  // Авторизация администратора
  loginAdmin: async (username, password) => {
    try {
      const response = await api.post('/auth/admin/login', { username, password });
      const { token, expires_at, is_admin } = response.data;
      
      // Сохраняем данные в localStorage
      localStorage.setItem('token', token);
      localStorage.setItem('expiresAt', expires_at);
      localStorage.setItem('isAdmin', is_admin);
      localStorage.setItem('user', JSON.stringify({ username, is_admin }));
      
      return response.data;
    } catch (error) {
      console.error('Ошибка при авторизации администратора:', error);
      throw error;
    }
  },
  
  // Авторизация кассира
  loginCashier: async (username, password) => {
    try {
      const response = await api.post('/auth/cashier/login', { username, password });
      const { token, expires_at, is_admin } = response.data;
      
      // Сохраняем данные в localStorage
      localStorage.setItem('token', token);
      localStorage.setItem('expiresAt', expires_at);
      localStorage.setItem('isAdmin', is_admin);
      localStorage.setItem('user', JSON.stringify({ username, is_admin }));
      
      return response.data;
    } catch (error) {
      console.error('Ошибка при авторизации кассира:', error);
      throw error;
    }
  },
  
  // Выход из системы
  logout: async () => {
    try {
      await api.post('/auth/logout');
    } catch (error) {
      console.error('Ошибка при выходе из системы:', error);
    } finally {
      // Удаляем данные из localStorage
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('expiresAt');
      localStorage.removeItem('isAdmin');
    }
  },
  
  // Получение текущего пользователя
  getCurrentUser: () => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      return JSON.parse(userStr);
    }
    return null;
  },
  
  // Проверка, авторизован ли пользователь
  isAuthenticated: () => {
    const token = localStorage.getItem('token');
    const expiresAt = localStorage.getItem('expiresAt');
    
    if (!token || !expiresAt) {
      return false;
    }
    
    // Проверяем, не истек ли токен
    const expirationDate = new Date(expiresAt);
    const now = new Date();
    
    return now < expirationDate;
  },
  
  // Проверка, является ли пользователь администратором
  isAdmin: () => {
    const isAdmin = localStorage.getItem('isAdmin');
    return isAdmin === 'true';
  },
  
  // Получение правильного пути для перенаправления после авторизации
  getRedirectPath: () => {
    const isAdmin = AuthService.isAdmin();
    return isAdmin ? '/admin' : '/cashier';
  },
  
  // Проверка, нужно ли перенаправить пользователя
  shouldRedirect: (currentPath) => {
    const isAuthenticated = AuthService.isAuthenticated();
    const isAdmin = AuthService.isAdmin();
    
    // Если пользователь не авторизован, не нужно перенаправлять
    if (!isAuthenticated) {
      return false;
    }
    
    // Если администратор находится на странице кассира, перенаправляем
    if (isAdmin && currentPath.startsWith('/cashier')) {
      return '/admin';
    }
    
    // Если кассир находится на странице администратора, перенаправляем
    if (!isAdmin && currentPath.startsWith('/admin')) {
      return '/cashier';
    }
    
    return false;
  },
  
  // Получение списка кассиров (только для администратора)
  getCashiers: async () => {
    try {
      const response = await api.get('/auth/cashiers');
      return response.data;
    } catch (error) {
      console.error('Ошибка при получении списка кассиров:', error);
      throw error;
    }
  },
  
  // Создание кассира (только для администратора)
  createCashier: async (username, password) => {
    try {
      const response = await api.post('/auth/cashiers', { username, password });
      return response.data;
    } catch (error) {
      console.error('Ошибка при создании кассира:', error);
      throw error;
    }
  },
  
  // Обновление кассира (только для администратора)
  updateCashier: async (id, data) => {
    try {
      const response = await api.put(`/auth/cashiers/${id}`, data);
      return response.data;
    } catch (error) {
      console.error('Ошибка при обновлении кассира:', error);
      throw error;
    }
  },
  
  // Удаление кассира (только для администратора)
  deleteCashier: async (id) => {
    try {
      const response = await api.delete(`/auth/cashiers/${id}`);
      return response.data;
    } catch (error) {
      console.error('Ошибка при удалении кассира:', error);
      throw error;
    }
  },
};

export default AuthService;
