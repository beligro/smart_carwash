/**
 * Утилиты для валидации и нормализации российских госномеров
 */

// Маппинг английских букв на русские для госномеров
const ENGLISH_TO_RUSSIAN = {
  'A': 'А', 'B': 'В', 'E': 'Е', 'K': 'К', 'M': 'М',
  'H': 'Н', 'O': 'О', 'P': 'Р', 'C': 'С', 'T': 'Т',
  'Y': 'У', 'X': 'Х',
  'a': 'А', 'b': 'В', 'e': 'Е', 'k': 'К', 'm': 'М',
  'h': 'Н', 'o': 'О', 'p': 'Р', 'c': 'С', 't': 'Т',
  'y': 'У', 'x': 'Х'
};

// Регулярное выражение для валидации российских госномеров
// Формат: О111ОО799 (буква + 3 цифры + 2 буквы + 2-3 цифры)
const RUSSIAN_LICENSE_PLATE_REGEX = /^[АВЕКМНОРСТУХ]\d{3}[АВЕКМНОРСТУХ]{2}\d{2,3}$/;

/**
 * Заменяет английские буквы на русские в госномере
 * @param {string} str - Строка для замены
 * @returns {string} - Строка с замененными буквами
 */
const replaceEnglishWithRussian = (str) => {
  if (!str || typeof str !== 'string') {
    return '';
  }
  
  return str
    .split('')
    .map(char => ENGLISH_TO_RUSSIAN[char] || char)
    .join('');
};

/**
 * Валидирует и нормализует российский госномер
 * Приводит к формату: заглавные русские буквы + цифры
 * 
 * @param {string} licensePlate - Госномер для валидации
 * @returns {Object} - Результат валидации { isValid: boolean, normalized: string, error: string }
 * 
 * @example
 * validateAndNormalizeLicensePlate("A123BC77") // { isValid: true, normalized: "А123ВС77", error: "" }
 * validateAndNormalizeLicensePlate("a123bc77") // { isValid: true, normalized: "А123ВС77", error: "" }
 * validateAndNormalizeLicensePlate("О111ОО799") // { isValid: true, normalized: "О111ОО799", error: "" }
 * validateAndNormalizeLicensePlate("o111oo799") // { isValid: true, normalized: "О111ОО799", error: "" }
 */
export const validateAndNormalizeLicensePlate = (licensePlate) => {
  try {
    // Проверяем входные данные
    if (!licensePlate || typeof licensePlate !== 'string') {
      return {
        isValid: false,
        normalized: '',
        error: 'Номер автомобиля не может быть пустым'
      };
    }

    // Очищаем от пробелов и дефисов
    const cleaned = licensePlate.replace(/[\s-]/g, '');

    // Приводим к верхнему регистру
    const upper = cleaned.toUpperCase();

    // Заменяем английские буквы на русские
    const normalized = replaceEnglishWithRussian(upper);

    // Валидируем формат
    if (!RUSSIAN_LICENSE_PLATE_REGEX.test(normalized)) {
      return {
        isValid: false,
        normalized: normalized,
        error: 'Неверный формат номера автомобиля. Используйте формат: О111ОО799'
      };
    }

    return {
      isValid: true,
      normalized: normalized,
      error: ''
    };
  } catch (error) {
    console.error('Ошибка валидации госномера:', error);
    return {
      isValid: false,
      normalized: '',
      error: 'Ошибка валидации госномера'
    };
  }
};

/**
 * Проверяет, является ли строка валидным российским госномером
 * без нормализации (для быстрой проверки)
 * 
 * @param {string} licensePlate - Госномер для проверки
 * @returns {boolean} - true если валидный, false если нет
 */
export const isValidLicensePlate = (licensePlate) => {
  try {
    if (!licensePlate || typeof licensePlate !== 'string') {
      return false;
    }
    return RUSSIAN_LICENSE_PLATE_REGEX.test(licensePlate);
  } catch (error) {
    console.error('Ошибка проверки госномера:', error);
    return false;
  }
};

/**
 * Нормализует госномер для поиска в базе данных
 * Используется в GetUserByCarNumber для поиска по нормализованному номеру
 * 
 * @param {string} licensePlate - Госномер для нормализации
 * @returns {string} - Нормализованный госномер
 */
export const normalizeLicensePlateForSearch = (licensePlate) => {
  try {
    if (!licensePlate || typeof licensePlate !== 'string') {
      return '';
    }

    // Очищаем и приводим к верхнему регистру
    const cleaned = licensePlate.replace(/[\s-]/g, '');
    const upper = cleaned.toUpperCase();

    // Заменяем английские буквы на русские
    return replaceEnglishWithRussian(upper);
  } catch (error) {
    console.error('Ошибка нормализации госномера:', error);
    return '';
  }
};

/**
 * Возвращает все возможные варианты госномера для поиска
 * Полезно для поиска в БД, где могут быть разные варианты написания
 * 
 * @param {string} licensePlate - Госномер для поиска вариантов
 * @returns {Array<string>} - Массив вариантов госномера
 */
export const getLicensePlateVariants = (licensePlate) => {
  try {
    const validation = validateAndNormalizeLicensePlate(licensePlate);
    if (!validation.isValid) {
      return [];
    }
    return [validation.normalized];
  } catch (error) {
    console.error('Ошибка получения вариантов госномера:', error);
    return [];
  }
};

/**
 * Форматирует госномер для отображения
 * Добавляет пробелы для лучшей читаемости
 * 
 * @param {string} licensePlate - Госномер для форматирования
 * @returns {string} - Отформатированный госномер
 */
export const formatLicensePlateForDisplay = (licensePlate) => {
  try {
    if (!licensePlate || typeof licensePlate !== 'string') {
      return '';
    }

    // Нормализуем номер
    const normalized = normalizeLicensePlateForSearch(licensePlate);
    
    // Форматируем: А123ВС77 -> А 123 ВС 77
    if (normalized.length >= 8) {
      return `${normalized.slice(0, 1)} ${normalized.slice(1, 4)} ${normalized.slice(4, 6)} ${normalized.slice(6)}`;
    }
    
    return normalized;
  } catch (error) {
    console.error('Ошибка форматирования госномера:', error);
    return licensePlate || '';
  }
};

/**
 * Получает примеры валидных госномеров для подсказки пользователю
 * 
 * @returns {Array<string>} - Массив примеров госномеров
 */
export const getLicensePlateExamples = () => {
  return [
    'А123ВС77',
    'О111ОО799',
    'М456ТУ123',
    'К789ХЕ45'
  ];
};

/**
 * Получает описание формата госномера для пользователя
 * 
 * @returns {string} - Описание формата
 */
export const getLicensePlateFormatDescription = () => {
  return 'Формат: буква + 3 цифры + 2 буквы + 2-3 цифры. Пример: А123ВС77';
};

/**
 * Проверяет, поддерживается ли буква в российских госномерах
 * 
 * @param {string} letter - Буква для проверки
 * @returns {boolean} - true если поддерживается, false если нет
 */
export const isSupportedLetter = (letter) => {
  if (!letter || typeof letter !== 'string' || letter.length !== 1) {
    return false;
  }
  
  const upperLetter = letter.toUpperCase();
  const supportedLetters = ['А', 'В', 'Е', 'К', 'М', 'Н', 'О', 'Р', 'С', 'Т', 'У', 'Х'];
  
  return supportedLetters.includes(upperLetter);
};

/**
 * Получает список поддерживаемых букв для российских госномеров
 * 
 * @returns {Array<string>} - Массив поддерживаемых букв
 */
export const getSupportedLetters = () => {
  return ['А', 'В', 'Е', 'К', 'М', 'Н', 'О', 'Р', 'С', 'Т', 'У', 'Х'];
};

export default {
  validateAndNormalizeLicensePlate,
  isValidLicensePlate,
  normalizeLicensePlateForSearch,
  getLicensePlateVariants,
  formatLicensePlateForDisplay,
  getLicensePlateExamples,
  getLicensePlateFormatDescription,
  isSupportedLetter,
  getSupportedLetters
};
