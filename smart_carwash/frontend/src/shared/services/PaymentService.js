import { api } from './ApiService';

// Сервис для работы с платежами
const PaymentService = {
  // Расчет стоимости услуги
  calculatePrice: async (serviceType, rentalTimeMinutes, withChemistry = false) => {
    try {
      const response = await api.post('/payments/calculate-price', {
        service_type: serviceType,
        rental_time_minutes: rentalTimeMinutes,
        with_chemistry: withChemistry,
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка расчета стоимости:', error);
      throw error;
    }
  },

  // Создание платежа за очередь
  createQueuePayment: async (userID, serviceType, rentalTimeMinutes, withChemistry = false) => {
    try {
      const response = await api.post('/payments/queue', null, {
        params: {
          user_id: userID,
          service_type: serviceType,
          rental_time_minutes: rentalTimeMinutes,
          with_chemistry: withChemistry,
        },
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка создания платежа за очередь:', error);
      throw error;
    }
  },

  // Создание платежа за продление сессии
  createSessionExtensionPayment: async (sessionID, extensionMinutes) => {
    try {
      const response = await api.post('/payments/session-extension', null, {
        params: {
          session_id: sessionID,
          extension_minutes: extensionMinutes,
        },
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка создания платежа за продление сессии:', error);
      throw error;
    }
  },

  // Получение статуса платежа
  getPaymentStatus: async (paymentID) => {
    try {
      const response = await api.get('/payments/status', {
        params: { id: paymentID },
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка получения статуса платежа:', error);
      throw error;
    }
  },

  // Получение платежа по ID
  getPayment: async (paymentID) => {
    try {
      const response = await api.get('/payments/by-id', {
        params: { id: paymentID },
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка получения платежа:', error);
      throw error;
    }
  },

  // Проверка платежа для сессии
  checkPaymentForSession: async (sessionID) => {
    try {
      const response = await api.get('/payments/session', {
        params: { session_id: sessionID },
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка проверки платежа для сессии:', error);
      throw error;
    }
  },

  // Создание возврата
  createRefund: async (paymentID, amountKopecks, type) => {
    try {
      const response = await api.post('/payments/refunds', {
        payment_id: paymentID,
        amount_kopecks: amountKopecks,
        type: type,
        idempotency_key: generateIdempotencyKey(),
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка создания возврата:', error);
      throw error;
    }
  },

  // Автоматический возврат
  processAutomaticRefund: async (sessionID) => {
    try {
      const response = await api.post('/payments/refunds/automatic', null, {
        params: { session_id: sessionID },
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка автоматического возврата:', error);
      throw error;
    }
  },

  // Полный возврат
  processFullRefund: async (paymentID) => {
    try {
      const response = await api.post('/payments/refunds/full', null, {
        params: { payment_id: paymentID },
      });
      return response.data;
    } catch (error) {
      console.error('Ошибка полного возврата:', error);
      throw error;
    }
  },
};

// Утилиты
function generateIdempotencyKey() {
  const timestamp = Date.now();
  const random = Math.random().toString(36).substring(2, 15);
  return `${timestamp}_${random}`;
}

// Форматирование суммы в копейках в рубли
export function formatAmount(kopecks) {
  const rubles = kopecks / 100;
  return rubles.toLocaleString('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: 0,
  });
}

// Устаревшие методы - оставляем для обратной совместимости
// Получение цены услуги в копейках (устарело - используйте calculatePrice)
export function getServicePrice(serviceType) {
  const prices = {
    wash: 5000,      // 50 рублей
    air_dry: 3000,   // 30 рублей
    vacuum: 2000,    // 20 рублей
  };
  
  return prices[serviceType] || 5000;
}

// Получение цены продления в копейках (устарело - используйте calculatePrice)
export function getExtensionPrice(minutes) {
  return minutes * 1000; // 10 рублей за минуту
}

export default PaymentService; 