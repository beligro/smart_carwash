/**
 * Получает текстовое представление статуса сессии
 * @param {string} status - Статус сессии
 * @returns {string} - Текстовое представление статуса
 */
export const getSessionStatusText = (status) => {
    switch (status) {
      case 'created':
        return 'В очереди';
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
        return 'Ожидание назначения бокса...';
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
  