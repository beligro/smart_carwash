/**
 * GuestApiService — API для гостевого режима.
 * Не требует JWT. Доступ к сессии через guest_token (cookie).
 * Сигнатуры методов совместимы с WebApiService, чтобы переиспользовать
 * общие компоненты (WashInfo и др.). sessionId/userId в аргументах
 * игнорируются — сессия определяется по guest_token из cookie.
 */

import axios from 'axios';

const COOKIE_NAME = 'guest_session_token';
const baseURL = process.env.REACT_APP_API_URL || '/api';

const guestApi = axios.create({
  baseURL,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
});

// --- Cookie helpers ---

export function getGuestToken() {
  const match = document.cookie.match(new RegExp('(^| )' + COOKIE_NAME + '=([^;]+)'));
  return match ? match[2] : null;
}

export function setGuestToken(token, hours = 24) {
  const expires = new Date(Date.now() + hours * 3600 * 1000).toUTCString();
  document.cookie = `${COOKIE_NAME}=${token}; expires=${expires}; path=/; SameSite=Lax`;
}

export function clearGuestToken() {
  document.cookie = `${COOKIE_NAME}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
}

function tokenOrThrow() {
  const t = getGuestToken();
  if (!t) throw new Error('Нет токена гостевой сессии');
  return t;
}

const GuestApiService = {
  // --- Создание сессии (без авторизации) ---
  createSession: async (data) => {
    const { v4: uuidv4 } = await import('uuid');
    const payload = {
      service_type: data.serviceType,
      with_chemistry: data.withChemistry || false,
      chemistry_time_minutes: data.chemistryTimeMinutes || 0,
      car_number: data.carNumber,
      car_number_country: data.carNumberCountry || 'RUS',
      rental_time_minutes: data.rentalTimeMinutes,
      idempotency_key: uuidv4(),
    };
    const res = await guestApi.post('/guest/sessions/with-payment', payload);
    if (res.data?.guest_token) setGuestToken(res.data.guest_token);
    return res.data;
  },

  // createSessionWithPayment — алиас под сигнатуру WebApiService (для WashInfo/booking)
  createSessionWithPayment: async (data) => {
    return GuestApiService.createSession({
      serviceType: data.serviceType || data.service_type,
      withChemistry: data.withChemistry ?? data.with_chemistry,
      chemistryTimeMinutes: data.chemistryTimeMinutes ?? data.chemistry_time_minutes,
      carNumber: data.carNumber || data.car_number,
      carNumberCountry: data.carNumberCountry || data.car_number_country,
      rentalTimeMinutes: data.rentalTimeMinutes || data.rental_time_minutes,
    });
  },

  // --- Чтение сессии (по токену; sessionId игнорируется) ---
  getSession: async (token) => {
    const t = token || tokenOrThrow();
    const res = await guestApi.get(`/guest/sessions/${t}`);
    return res.data;
  },

  getSessionById: async (_sessionId) => {
    const res = await guestApi.get(`/guest/sessions/${tokenOrThrow()}`);
    return res.data; // { session, payment }
  },

  getUserSession: async () => {
    try {
      const res = await guestApi.get(`/guest/sessions/${tokenOrThrow()}`);
      return res.data;
    } catch (e) {
      if (e.response?.status === 404) return { session: null };
      throw e;
    }
  },

  getUserSessionForPayment: async (_userId) => {
    const res = await guestApi.get(`/guest/sessions/${tokenOrThrow()}`);
    return res.data;
  },

  getSessionPayments: async (_sessionId) => {
    const res = await guestApi.get(`/guest/sessions/${tokenOrThrow()}/payments`);
    return res.data; // { main_payment, extension_payments }
  },

  // --- Действия (по токену) ---
  startSession: async (_sessionId) => {
    const res = await guestApi.post(`/guest/sessions/${tokenOrThrow()}/start`, {});
    return res.data;
  },

  completeSession: async (_sessionId) => {
    const res = await guestApi.post(`/guest/sessions/${tokenOrThrow()}/complete`, {});
    return res.data;
  },

  cancelSession: async (_sessionId) => {
    const res = await guestApi.post(`/guest/sessions/${tokenOrThrow()}/cancel`, {});
    return res.data;
  },

  enableChemistry: async (_sessionId) => {
    const res = await guestApi.post(`/guest/sessions/${tokenOrThrow()}/enable-chemistry`, {});
    return res.data;
  },

  extendSessionWithPayment: async (_sessionId, extensionTimeMinutes, extensionChemistryTimeMinutes = 0) => {
    const res = await guestApi.post(`/guest/sessions/${tokenOrThrow()}/extend`, {
      extension_time_minutes: extensionTimeMinutes,
      extension_chemistry_time_minutes: extensionChemistryTimeMinutes,
    });
    return res.data;
  },

  // --- Публичные справочники (без токена) ---
  getQueueStatus: async () => {
    const res = await guestApi.get('/queue-status');
    return res.data;
  },

  getCarwashStatus: async () => {
    const res = await guestApi.get('/carwash/status');
    return res.data;
  },

  calculatePrice: async (data) => {
    const res = await guestApi.post('/payments/calculate-price', {
      service_type: data.serviceType || data.service_type,
      with_chemistry: data.withChemistry ?? data.with_chemistry ?? false,
      chemistry_time_minutes: data.chemistryTimeMinutes ?? data.chemistry_time_minutes ?? 0,
      rental_time_minutes: data.rentalTimeMinutes || data.rental_time_minutes,
    });
    return res.data;
  },

  getAvailableRentalTimes: async (serviceType) => {
    try {
      const res = await guestApi.get(`/settings/rental-times?service_type=${serviceType}`);
      return res.data;
    } catch {
      return { availableTimes: [15, 20, 30, 60, 90] };
    }
  },

  getAvailableChemistryTimes: async (serviceType) => {
    try {
      const res = await guestApi.get(`/settings/available-chemistry-times?service_type=${serviceType}`);
      return res.data;
    } catch {
      return { available_chemistry_times: [3, 4, 10] };
    }
  },
};

export default GuestApiService;
