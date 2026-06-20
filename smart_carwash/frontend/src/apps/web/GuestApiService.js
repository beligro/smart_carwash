/**
 * GuestApiService — API для гостевого режима.
 * Не требует JWT. Доступ к сессии через guest_token, хранящийся в cookie.
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

// --- API methods ---

const GuestApiService = {
  /**
   * Создать сессию без авторизации.
   * Возвращает { session, payment, guest_token }.
   * После успешного вызова сохраняет guest_token в cookie.
   */
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
    if (res.data?.guest_token) {
      setGuestToken(res.data.guest_token);
    }
    return res.data;
  },

  /**
   * Получить статус сессии по guest_token.
   * Возвращает { session, payment }.
   */
  getSession: async (token) => {
    const t = token || getGuestToken();
    if (!t) throw new Error('Нет токена гостевой сессии');
    const res = await guestApi.get(`/guest/sessions/${t}`);
    return res.data;
  },

  /**
   * Продлить сессию.
   * Возвращает { session, payment } с новым платёжным URL.
   */
  extendSession: async (extensionTimeMinutes, extensionChemistryTimeMinutes = 0, token) => {
    const t = token || getGuestToken();
    if (!t) throw new Error('Нет токена гостевой сессии');
    const res = await guestApi.post(`/guest/sessions/${t}/extend`, {
      extension_time_minutes: extensionTimeMinutes,
      extension_chemistry_time_minutes: extensionChemistryTimeMinutes,
    });
    return res.data;
  },

  /**
   * Получить статус очереди (публичный endpoint — без авторизации).
   */
  getQueueStatus: async () => {
    const res = await guestApi.get('/queue-status');
    return res.data;
  },

  /**
   * Получить статус мойки (публичный).
   */
  getCarwashStatus: async () => {
    const res = await guestApi.get('/carwash/status');
    return res.data;
  },

  /**
   * Рассчитать стоимость услуги.
   */
  calculatePrice: async (data) => {
    const res = await guestApi.post('/payments/calculate-price', {
      service_type: data.serviceType,
      with_chemistry: data.withChemistry || false,
      chemistry_time_minutes: data.chemistryTimeMinutes || 0,
      rental_time_minutes: data.rentalTimeMinutes,
    });
    return res.data;
  },

  /**
   * Получить доступные времена аренды.
   */
  getAvailableRentalTimes: async (serviceType) => {
    try {
      const res = await guestApi.get(`/settings/rental-times?service_type=${serviceType}`);
      return res.data;
    } catch {
      return { availableTimes: [15, 20, 30, 60, 90] };
    }
  },

  /**
   * Получить доступные времена химии.
   */
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
