/**
 * WebApiService - API для веб-версии (JWT, префикс /web и /auth/web).
 * Токен хранится в localStorage под ключом web_token.
 */

import axios from 'axios';
import { v4 as uuidv4 } from 'uuid';
import { toSnakeCase } from '../utils/snakeCase';

const baseURL = process.env.REACT_APP_API_URL || '/api';

const webApi = axios.create({
  baseURL,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
});

webApi.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('web_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (err) => Promise.reject(err)
);

webApi.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('web_token');
      localStorage.removeItem('web_user');
      if (window.location.pathname.startsWith('/web') && !window.location.pathname.startsWith('/web/login')) {
        window.location.href = '/web/login';
      }
    }
    return Promise.reject(err);
  }
);

const WebApiService = {
  login: async (email, password) => {
    const data = toSnakeCase({ email, password });
    const res = await axios.post(`${baseURL}/auth/web/login`, data, {
      headers: { 'Content-Type': 'application/json' },
      timeout: 10000,
    });
    return res.data;
  },

  registerSendCode: async (email, password, passwordConfirm) => {
    const data = toSnakeCase({ email, password, password_confirm: passwordConfirm });
    const res = await axios.post(`${baseURL}/auth/web/register/send-code`, data, {
      headers: { 'Content-Type': 'application/json' },
      timeout: 10000,
    });
    return res.data;
  },

  registerVerify: async (email, code, password, passwordConfirm) => {
    const data = toSnakeCase({ email, code, password, password_confirm: passwordConfirm });
    const res = await axios.post(`${baseURL}/auth/web/register/verify`, data, {
      headers: { 'Content-Type': 'application/json' },
      timeout: 10000,
    });
    return res.data;
  },

  logout: async () => {
    const token = localStorage.getItem('web_token');
    try {
      await axios.post(
        `${baseURL}/auth/web/logout`,
        {},
        { headers: { Authorization: `Bearer ${token}` }, timeout: 5000 }
      );
    } finally {
      localStorage.removeItem('web_token');
      localStorage.removeItem('web_user');
    }
  },

  changePasswordSendCode: async (email) => {
    const data = toSnakeCase({ email });
    const res = await webApi.post('/auth/web/change-password/send-code', data);
    return res.data;
  },

  changePasswordConfirm: async (email, code, newPassword, newPasswordConfirm) => {
    const data = toSnakeCase({
      email,
      code,
      new_password: newPassword,
      new_password_confirm: newPasswordConfirm,
    });
    const res = await webApi.post('/auth/web/change-password/confirm', data);
    return res.data;
  },

  getMe: async () => {
    const res = await webApi.get('/web/me');
    return res.data;
  },

  getQueueStatus: async () => {
    const res = await webApi.get('/queue-status');
    return res.data;
  },

  getCarwashStatus: async () => {
    const res = await axios.get(`${baseURL}/carwash/status`, { timeout: 10000 });
    return res.data;
  },

  getUserSession: async () => {
    try {
      const res = await webApi.get('/web/sessions');
      return res.data;
    } catch (e) {
      if (e.response?.status === 404) return { session: null };
      throw e;
    }
  },

  getUserSessionForPayment: async () => {
    try {
      const res = await webApi.get('/web/sessions/for-payment');
      return res.data;
    } catch (e) {
      if (e.response?.status === 404) return { session: null };
      throw e;
    }
  },

  checkActiveSession: async () => {
    const res = await webApi.get('/web/sessions/check-active');
    return res.data;
  },

  getSessionById: async (sessionId) => {
    const res = await webApi.get(`/web/sessions/by-id?session_id=${sessionId}`);
    return res.data;
  },

  createSessionWithPayment: async (data) => {
    const payload = { ...data, idempotencyKey: uuidv4() };
    const snake = toSnakeCase(payload);
    const res = await webApi.post('/web/sessions/with-payment', snake);
    return res.data;
  },

  startSession: async (sessionId) => {
    const res = await webApi.post('/web/sessions/start', { session_id: sessionId });
    return res.data;
  },

  completeSession: async (sessionId) => {
    const res = await webApi.post('/web/sessions/complete', { session_id: sessionId });
    return res.data;
  },

  extendSessionWithPayment: async (sessionId, extensionTimeMinutes, extensionChemistryTimeMinutes = 0) => {
    const body = { session_id: sessionId, extension_time_minutes: extensionTimeMinutes };
    if (extensionChemistryTimeMinutes > 0) {
      body.extension_chemistry_time_minutes = extensionChemistryTimeMinutes;
    }
    const res = await webApi.post('/web/sessions/extend-with-payment', body);
    return res.data;
  },

  getSessionPayments: async (sessionId) => {
    const res = await webApi.get(`/web/sessions/payments?session_id=${sessionId}`);
    return res.data;
  },

  getUserSessionHistory: async (limit = 5, offset = 0) => {
    const res = await webApi.get(`/web/sessions/history?limit=${limit}&offset=${offset}`);
    return res.data;
  },

  cancelSession: async (sessionId) => {
    const res = await webApi.post('/web/sessions/cancel', { session_id: sessionId });
    return res.data;
  },

  enableChemistry: async (sessionId) => {
    const res = await webApi.post('/web/sessions/enable-chemistry', { session_id: sessionId });
    return res.data;
  },

  getAvailableRentalTimes: async (serviceType) => {
    try {
      const res = await webApi.get(`/settings/rental-times?service_type=${serviceType}`);
      return res.data;
    } catch {
      return { availableTimes: [15, 20, 30, 60, 90] };
    }
  },

  getAvailableChemistryTimes: async (serviceType) => {
    try {
      const res = await webApi.get(`/settings/available-chemistry-times?service_type=${serviceType}`);
      return res.data;
    } catch {
      return { available_chemistry_times: [3, 4, 10] };
    }
  },

  calculatePrice: async (data) => {
    const snake = toSnakeCase(data);
    const res = await webApi.post('/payments/calculate-price', snake);
    return res.data;
  },

  updateCarNumber: async (userId, carNumber, carNumberCountry = 'RUS') => {
    const res = await webApi.put('/users/car-number', toSnakeCase({ userId, carNumber, carNumberCountry }));
    return res.data;
  },

  linkTelegramRequest: async () => {
    const res = await webApi.post('/web/link-telegram/request');
    return res.data;
  },

  getPaymentStatus: async (paymentId) => {
    const res = await webApi.get(`/web/payments/status?payment_id=${paymentId}`);
    return res.data;
  },

  createNewPayment: async (sessionId, amount, currency) => {
    const res = await webApi.post('/web/payments/create', toSnakeCase({ sessionId, amount, currency }));
    return res.data;
  },
};

export default WebApiService;
