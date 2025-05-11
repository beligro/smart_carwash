import axios from 'axios';

// Create axios instance with base URL from environment
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'https://localhost/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// API endpoints
const endpoints = {
  // Carwash endpoints
  carwash: {
    getInfo: () => api.get('/carwash/info'),
    getBox: (boxId) => api.get(`/carwash/boxes/${boxId}`),
    updateBoxStatus: (boxId, status) => api.patch(`/carwash/boxes/${boxId}`, { status }),
  },
  
  // User endpoints
  users: {
    create: (userData) => api.post('/users', userData),
    getByTelegramId: (telegramId) => api.get(`/users/${telegramId}`),
    ensureExists: (userData) => api.post('/users/ensure', userData),
  },
};

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // You can add auth tokens or other headers here if needed
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    // Return just the data part of the response
    return response.data;
  },
  (error) => {
    // Handle errors
    if (error.response) {
      // The request was made and the server responded with a status code
      // that falls out of the range of 2xx
      console.error('API Error:', error.response.data);
    } else if (error.request) {
      // The request was made but no response was received
      console.error('Network Error:', error.request);
    } else {
      // Something happened in setting up the request that triggered an Error
      console.error('Request Error:', error.message);
    }
    
    return Promise.reject(error);
  }
);

export default endpoints;
