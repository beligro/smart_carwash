/**
 * Получает текстовое представление статуса сессии
 * @param {string} status - Статус сессии
 * @returns {string} - Текстовое представление статуса
 */
export const getSessionStatusText = (status) => {
    switch (status) {
      case 'created':
        return 'Создана';
      case 'in_queue':
        return 'В очереди';
      case 'payment_failed':
        return 'Ошибка оплаты';
      case 'assigned':
        return 'Назначена';
      case 'active':
        return 'Активна';
      case 'complete':
        return 'Завершена';
      case 'canceled':
        return 'Отменена';
      case 'expired':
        return 'Истекла';
      default:
        return 'Неизвестно';
    }
  };
  
  /**
   * Получает текстовое представление статуса бокса
   * @param {string} status - Статус бокса
   * @returns {string} - Текстовое представление статуса
   */
  export const getBoxStatusText = (status) => {
    switch (status) {
      case 'free':
        return 'Свободен';
      case 'reserved':
        return 'Зарезервирован';
      case 'busy':
        return 'Занят';
      case 'maintenance':
        return 'На обслуживании';
      default:
        return 'Неизвестно';
    }
  };
  
  /**
 * Получает описание статуса сессии для отображения пользователю
 * @param {string} status - Статус сессии
 * @returns {string} - Описание статуса
 */
export const getSessionStatusDescription = (status) => {
    switch (status) {
      case 'created':
        return 'Сессия создана, ожидание оплаты...';
      case 'in_queue':
        return 'Оплачено, ожидание назначения бокса...';
      case 'payment_failed':
        return 'Ошибка оплаты, попробуйте еще раз';
      case 'assigned':
        return 'Бокс назначен, ожидание клиента';
      case 'active':
        return 'Мойка в процессе';
      case 'complete':
        return 'Мойка завершена';
      case 'canceled':
        return 'Сессия отменена';
      case 'expired':
        return 'Время сессии истекло';
      default:
        return 'Неизвестный статус';
    }
  };
  
  /**
   * Получает описание типа услуги для отображения пользователю
   * @param {string} serviceType - Тип услуги
   * @returns {string} - Описание типа услуги
   */
  export const getServiceTypeDescription = (serviceType) => {
  // Добавляем логирование для отладки
  if (!serviceType) {
    console.warn('getServiceTypeDescription: serviceType is null or undefined');
    return 'Тип услуги не указан';
  }
  
  console.log('getServiceTypeDescription: serviceType =', serviceType);
  
  switch (serviceType) {
    case 'wash':
      return 'Мойка';
    case 'air_dry':
      return 'Обдув воздухом';
    case 'vacuum':
      return 'Пылесос';
    default:
      console.warn('getServiceTypeDescription: unknown serviceType =', serviceType);
      // Возвращаем исходное значение вместо "Неизвестная услуга"
      return serviceType;
  }
};
