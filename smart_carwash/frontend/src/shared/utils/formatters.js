import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

/**
 * Форматирует дату в читаемый формат
 * @param {string} dateString - Строка с датой
 * @returns {string} - Отформатированная дата
 */
export const formatDate = (dateString) => {
  if (!dateString) {
    console.warn('formatDate: dateString is null or undefined');
    return 'Дата не указана';
  }
  
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      console.warn('formatDate: invalid date string:', dateString);
      // Возвращаем исходную строку вместо "Некорректная дата"
      return dateString;
    }
    return format(date, 'dd MMMM yyyy, HH:mm', { locale: ru });
  } catch (error) {
    console.error('Ошибка форматирования даты:', error, 'dateString:', dateString);
    // Возвращаем исходную строку вместо "Ошибка форматирования даты"
    return dateString;
  }
};

/**
 * Форматирует время в формат MM:SS
 * @param {number} seconds - Количество секунд
 * @returns {string} - Отформатированное время
 */
export const formatTime = (seconds) => {
  if (seconds === null || seconds === undefined) return '00:00';
  
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  
  return `${minutes}:${remainingSeconds < 10 ? '0' : ''}${remainingSeconds}`;
};
