/**
 * Утилиты для работы с датами и временными зонами
 * Решает проблему конвертации локального времени в UTC+0 для корректной работы с бэкендом
 */

/**
 * Конвертирует локальную дату и время в UTC ISO строку
 * @param {string} dateString - дата в формате YYYY-MM-DD
 * @param {string} timeString - время в формате HH:MM (опционально)
 * @returns {string} - ISO строка в UTC
 */
export function convertLocalDateTimeToUTC(dateString, timeString = '00:00') {
  if (!dateString) return null;
  
  // Создаем дату в локальной временной зоне
  const [year, month, day] = dateString.split('-').map(Number);
  const [hours, minutes] = timeString.split(':').map(Number);
  
  const localDate = new Date(year, month - 1, day, hours || 0, minutes || 0, 0, 0);
  
  // Возвращаем ISO строку в UTC
  return localDate.toISOString();
}

/**
 * Получает начало дня (00:00:00) в UTC для локальной даты
 * @param {Date|string} date - дата (объект Date или строка YYYY-MM-DD)
 * @returns {string} - ISO строка начала дня в UTC
 */
export function getStartOfDayUTC(date) {
  const dateObj = typeof date === 'string' ? new Date(date + 'T00:00:00') : new Date(date);
  const year = dateObj.getFullYear();
  const month = dateObj.getMonth();
  const day = dateObj.getDate();
  
  // Создаем начало дня в локальной зоне
  const startOfDay = new Date(year, month, day, 0, 0, 0, 0);
  
  return startOfDay.toISOString();
}

/**
 * Получает конец дня (23:59:59) в UTC для локальной даты
 * @param {Date|string} date - дата (объект Date или строка YYYY-MM-DD)
 * @returns {string} - ISO строка конца дня в UTC
 */
export function getEndOfDayUTC(date) {
  const dateObj = typeof date === 'string' ? new Date(date + 'T00:00:00') : new Date(date);
  const year = dateObj.getFullYear();
  const month = dateObj.getMonth();
  const day = dateObj.getDate();
  
  // Создаем конец дня в локальной зоне
  const endOfDay = new Date(year, month, day, 23, 59, 59, 999);
  
  return endOfDay.toISOString();
}

/**
 * Конвертирует datetime-local значение в UTC ISO строку
 * @param {string} datetimeLocal - значение из input[type="datetime-local"]
 * @returns {string} - ISO строка в UTC
 */
export function convertDateTimeLocalToUTC(datetimeLocal) {
  if (!datetimeLocal) return null;
  
  // datetime-local возвращает значение в формате YYYY-MM-DDTHH:MM
  // Создаем Date объект в локальной зоне
  const localDate = new Date(datetimeLocal);
  
  return localDate.toISOString();
}

/**
 * Конвертирует UTC ISO строку в datetime-local формат для input
 * @param {string} utcISOString - ISO строка в UTC
 * @returns {string} - значение для input[type="datetime-local"]
 */
export function convertUTCToDateTimeLocal(utcISOString) {
  if (!utcISOString) return '';
  
  const date = new Date(utcISOString);
  
  // Форматируем в YYYY-MM-DDTHH:MM для datetime-local
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  
  return `${year}-${month}-${day}T${hours}:${minutes}`;
}

/**
 * Получает быстрые фильтры с правильной конвертацией в UTC
 * @param {string} filterType - тип фильтра (today, yesterday, this_week, etc.)
 * @returns {Object} - объект с dateFrom и dateTo в UTC
 */
export function getQuickFilterDates(filterType) {
  const today = new Date();
  
  switch (filterType) {
    case 'today':
      return {
        dateFrom: getStartOfDayUTC(today),
        dateTo: getEndOfDayUTC(today)
      };
      
    case 'yesterday':
      const yesterday = new Date(today);
      yesterday.setDate(yesterday.getDate() - 1);
      return {
        dateFrom: getStartOfDayUTC(yesterday),
        dateTo: getEndOfDayUTC(yesterday)
      };
      
    case 'this_week':
      const startOfWeek = new Date(today);
      startOfWeek.setDate(today.getDate() - today.getDay());
      return {
        dateFrom: getStartOfDayUTC(startOfWeek),
        dateTo: getEndOfDayUTC(today)
      };
      
    case 'last_week':
      const lastWeekStart = new Date(today);
      lastWeekStart.setDate(today.getDate() - today.getDay() - 7);
      const lastWeekEnd = new Date(lastWeekStart);
      lastWeekEnd.setDate(lastWeekStart.getDate() + 6);
      return {
        dateFrom: getStartOfDayUTC(lastWeekStart),
        dateTo: getEndOfDayUTC(lastWeekEnd)
      };
      
    case 'this_month':
      const startOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      return {
        dateFrom: getStartOfDayUTC(startOfMonth),
        dateTo: getEndOfDayUTC(today)
      };
      
    case 'last_month':
      const lastMonth = new Date(today.getFullYear(), today.getMonth() - 1, 1);
      const lastMonthEnd = new Date(today.getFullYear(), today.getMonth(), 0);
      return {
        dateFrom: getStartOfDayUTC(lastMonth),
        dateTo: getEndOfDayUTC(lastMonthEnd)
      };
      
    case 'this_year':
      const startOfYear = new Date(today.getFullYear(), 0, 1);
      return {
        dateFrom: getStartOfDayUTC(startOfYear),
        dateTo: getEndOfDayUTC(today)
      };
      
    case 'last_year':
      const lastYearStart = new Date(today.getFullYear() - 1, 0, 1);
      const lastYearEnd = new Date(today.getFullYear() - 1, 11, 31);
      return {
        dateFrom: getStartOfDayUTC(lastYearStart),
        dateTo: getEndOfDayUTC(lastYearEnd)
      };
      
    default:
      return { dateFrom: null, dateTo: null };
  }
}

/**
 * Форматирует дату для отображения в локальной временной зоне
 * @param {string} utcISOString - ISO строка в UTC
 * @param {Object} options - опции форматирования
 * @returns {string} - отформатированная дата
 */
export function formatDateForDisplay(utcISOString, options = {}) {
  if (!utcISOString) return '-';
  
  const date = new Date(utcISOString);
  const defaultOptions = {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    timeZoneName: 'short'
  };
  
  return date.toLocaleString('ru-RU', { ...defaultOptions, ...options });
}

/**
 * Получает текущую временную зону пользователя
 * @returns {string} - временная зона (например, "Europe/Moscow")
 */
export function getUserTimezone() {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

/**
 * Проверяет, является ли строка валидной датой
 * @param {string} dateString - строка даты
 * @returns {boolean} - true если дата валидна
 */
export function isValidDate(dateString) {
  if (!dateString) return false;
  const date = new Date(dateString);
  return !isNaN(date.getTime());
}

