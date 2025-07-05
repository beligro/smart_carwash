import _ from 'lodash';

/**
 * Рекурсивно преобразует все ключи объекта в snake_case
 * Поддерживает вложенные объекты, массивы и примитивы
 * 
 * @param {any} obj - объект для преобразования
 * @returns {any} - объект с ключами в snake_case
 * 
 * @example
 * toSnakeCase({ userInfo: { firstName: 'John', lastName: 'Doe' } })
 * // => { user_info: { first_name: 'John', last_name: 'Doe' } }
 * 
 * @example
 * toSnakeCase({ serviceType: 'wash', rentalTimeMinutes: 30 })
 * // => { service_type: 'wash', rental_time_minutes: 30 }
 */
export function toSnakeCase(obj) {
  // Если это массив - обрабатываем каждый элемент
  if (Array.isArray(obj)) {
    return obj.map(toSnakeCase);
  }
  
  // Если это объект (но не null, не Date, не примитив)
  if (obj && typeof obj === 'object' && obj.constructor === Object) {
    return Object.fromEntries(
      Object.entries(obj).map(([key, value]) => [
        _.snakeCase(key), 
        toSnakeCase(value)
      ])
    );
  }
  
  // Для примитивов, null, undefined, Date и других типов - возвращаем как есть
  return obj;
}

/**
 * Рекурсивно преобразует все ключи объекта из snake_case в camelCase
 * Поддерживает вложенные объекты, массивы и примитивы
 * 
 * @param {any} obj - объект для преобразования
 * @returns {any} - объект с ключами в camelCase
 * 
 * @example
 * fromSnakeCase({ user_info: { first_name: 'John', last_name: 'Doe' } })
 * // => { userInfo: { firstName: 'John', lastName: 'Doe' } }
 * 
 * @example
 * fromSnakeCase({ service_type: 'wash', rental_time_minutes: 30 })
 * // => { serviceType: 'wash', rentalTimeMinutes: 30 }
 */
export function fromSnakeCase(obj) {
  // Если это массив - обрабатываем каждый элемент
  if (Array.isArray(obj)) {
    return obj.map(fromSnakeCase);
  }
  
  // Если это объект (но не null, не Date, не примитив)
  if (obj && typeof obj === 'object' && obj.constructor === Object) {
    return Object.fromEntries(
      Object.entries(obj).map(([key, value]) => [
        _.camelCase(key), 
        fromSnakeCase(value)
      ])
    );
  }
  
  // Для примитивов, null, undefined, Date и других типов - возвращаем как есть
  return obj;
}

/**
 * Преобразует объект фильтров в query-строку с snake_case ключами
 * 
 * @param {Object} filters - объект фильтров
 * @returns {string} - query-строка
 * 
 * @example
 * toSnakeCaseQuery({ dateFrom: '2024-01-01', serviceType: 'wash' })
 * // => 'date_from=2024-01-01&service_type=wash'
 */
export function toSnakeCaseQuery(filters = {}) {
  const snakeFilters = toSnakeCase(filters);
  const params = new URLSearchParams();
  
  Object.keys(snakeFilters).forEach(key => {
    if (snakeFilters[key] !== undefined && snakeFilters[key] !== '') {
      params.append(key, snakeFilters[key]);
    }
  });
  
  return params.toString();
}

/**
 * Преобразует объект в JSON с snake_case ключами
 * 
 * @param {Object} obj - объект для преобразования
 * @returns {string} - JSON строка
 */
export function toSnakeCaseJSON(obj) {
  return JSON.stringify(toSnakeCase(obj));
} 