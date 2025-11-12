/**
 * Утилиты для валидации и нормализации госномеров разных стран
 */

// Маппинг русских букв на английские для госномеров
const RUSSIAN_TO_ENGLISH = {
  'А': 'A', 'В': 'B', 'Е': 'E', 'К': 'K', 'М': 'M',
  'Н': 'H', 'О': 'O', 'Р': 'P', 'С': 'C', 'Т': 'T',
  'У': 'Y', 'Х': 'X',
  'а': 'A', 'в': 'B', 'е': 'E', 'к': 'K', 'м': 'M',
  'н': 'H', 'о': 'O', 'р': 'P', 'с': 'C', 'т': 'T',
  'у': 'Y', 'х': 'X'
};

// Конфигурация стран и их форматов госномеров
const LICENSE_PLATE_COUNTRIES = {
    RUS: {
        // Россия: A123BC77
        pattern: /^[ABEKMHOPCTYX]\d{3}[ABEKMHOPCTYX]{2}\d{2,3}$/i,
        name: 'Россия',
        placeholder: 'A123BC77',
        examples: ['A123BC77', 'O111OO799', 'M456TY123']
    },
    KAZ: {
        // Казахстан: поддержка форматов 143ABA12, 927PK06, A000AAA
        pattern: /^(?:\d{3}[A-Z]{2,3}\d{2}|[A-Z]\d{3}[A-Z]{3})$/i,
        name: 'Казахстан',
        placeholder: '123ABC01',
        examples: ['143ABA12', '927PK06', 'A000AAA']
    },
    KGZ: {
        // Кыргызстан: поддержка всех форматов номеров
        pattern: /^(?:\d{3}[A-Z]{2}|\d{4}[A-Z]{2}|[A-Z]\d{4}[A-Z]{1,2}|\d{3}[A-Z]{3}|\d{4}[A-Z]{3}|\d{2}[A-Z]{2}\d{3}[A-Z]{3})$/i,
        name: 'Кыргызстан',
        placeholder: '1234AB',
        examples: ['123AB', '1234AB', 'B1234A', 'B1234AB', 'E1233E', '123ABC', '1234ABC', '01KG123ABC']
    },
    UZB: {
        // Узбекистан: поддержка форматов 01A123BC, 10AA1234
        pattern: /^(?:\d{2}[A-Z]\d{3}[A-Z]{2}|\d{2}[A-Z]{2}\d{4})$/i,
        name: 'Узбекистан',
        placeholder: '01A123BC',
        examples: ['01A123BC', '10AA1234']
    },
    TJK: {
        // Таджикистан: 1234AB01
        pattern: /^(\d{4})([A-Z]{2})(\d{2})$/i,
        name: 'Таджикистан',
        placeholder: '1234AB01',
        examples: ['1234AB01', '5678CD02']
    },
    ARM: {
        // Армения: поддержка форматов 123AB45, 01AA123
        pattern: /^(?:\d{3}[A-Z]{2}\d{2}|\d{2}[A-Z]{2}\d{3})$/i,
        name: 'Армения',
        placeholder: '123AB45',
        examples: ['123AB45', '01AA123']
    },
    BLR: {
        // Беларусь: поддержка форматов 0000AA-7, AA00AA00, AA1234-7 (дефис удаляется при валидации)
        pattern: /^(?:\d{4}[A-Z]{2}\d|[A-Z]{2}\d{2}[A-Z]{2}\d{2}|[A-Z]{2}\d{4}\d)$/i,
        name: 'Беларусь',
        placeholder: '1234AB1',
        examples: ['0000AA7', 'AA00AA00', 'AA12347']
    }
};

// Порядок стран для отображения
const COUNTRY_ORDER = ['RUS', 'KAZ', 'KGZ', 'UZB', 'TJK', 'ARM', 'BLR'];

/**
 * Заменяет русские буквы на английские в госномере
 * @param {string} str - Строка для замены
 * @returns {string} - Строка с замененными буквами
 */
const replaceRussianWithEnglish = (str) => {
  if (!str || typeof str !== 'string') {
    return '';
  }
  
  return str
    .split('')
    .map(char => RUSSIAN_TO_ENGLISH[char] || char)
    .join('');
};

/**
 * Валидирует и нормализует госномер для указанной страны
 * Приводит к формату: заглавные английские буквы + цифры
 * 
 * @param {string} licensePlate - Госномер для валидации
 * @param {string} country - Код страны (RUS, KAZ, KGZ, UZB, TJK, ARM, BLR)
 * @returns {Object} - Результат валидации { isValid: boolean, normalized: string, error: string }
 * 
 * @example
 * validateAndNormalizeLicensePlate("А123ВС77", "RUS") // { isValid: true, normalized: "A123BC77", error: "" }
 * validateAndNormalizeLicensePlate("123АВС01", "KAZ") // { isValid: true, normalized: "123ABC01", error: "" }
 */
export const validateAndNormalizeLicensePlate = (licensePlate, country = 'RUS') => {
  try {
    // Проверяем входные данные
    if (!licensePlate || typeof licensePlate !== 'string') {
      return {
        isValid: false,
        normalized: '',
        error: 'Номер автомобиля не может быть пустым'
      };
    }

    // Получаем конфигурацию страны
    const countryConfig = LICENSE_PLATE_COUNTRIES[country];
    if (!countryConfig) {
      return {
        isValid: false,
        normalized: '',
        error: 'Неподдерживаемая страна'
      };
    }

    // Очищаем от пробелов и дефисов
    const cleaned = licensePlate.replace(/[\s-]/g, '');

    // Приводим к верхнему регистру
    const upper = cleaned.toUpperCase();

    // Заменяем русские буквы на английские
    const normalized = replaceRussianWithEnglish(upper);

    // Валидируем формат для выбранной страны
    if (!countryConfig.pattern.test(normalized)) {
      return {
        isValid: false,
        normalized: normalized,
        error: `Неверный формат номера автомобиля для ${countryConfig.name}. Используйте формат: ${countryConfig.placeholder}`
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
 * Проверяет, является ли строка валидным госномером для указанной страны
 * без нормализации (для быстрой проверки)
 * 
 * @param {string} licensePlate - Госномер для проверки
 * @param {string} country - Код страны (RUS, KAZ, KGZ, UZB, TJK, ARM, BLR)
 * @returns {boolean} - true если валидный, false если нет
 */
export const isValidLicensePlate = (licensePlate, country = 'RUS') => {
  try {
    if (!licensePlate || typeof licensePlate !== 'string') {
      return false;
    }
    
    const countryConfig = LICENSE_PLATE_COUNTRIES[country];
    if (!countryConfig) {
      return false;
    }
    
    // Нормализуем номер перед проверкой (удаляем пробелы и дефисы, заменяем русские буквы)
    const cleaned = licensePlate.replace(/[\s-]/g, '');
    const upper = cleaned.toUpperCase();
    const normalized = replaceRussianWithEnglish(upper);
    
    return countryConfig.pattern.test(normalized);
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

    // Заменяем русские буквы на английские
    return replaceRussianWithEnglish(upper);
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
export const getLicensePlateVariants = (licensePlate, country = 'RUS') => {
  try {
    const validation = validateAndNormalizeLicensePlate(licensePlate, country);
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
    
    // Форматируем: A123BC77 -> A 123 BC 77
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
 * Получает список всех поддерживаемых стран
 * @returns {Array<Object>} - Массив объектов стран с кодом, названием и примерами
 */
export const getSupportedCountries = () => {
  return COUNTRY_ORDER.map(countryCode => ({
    code: countryCode,
    name: LICENSE_PLATE_COUNTRIES[countryCode].name,
    placeholder: LICENSE_PLATE_COUNTRIES[countryCode].placeholder,
    examples: LICENSE_PLATE_COUNTRIES[countryCode].examples
  }));
};

/**
 * Получает конфигурацию страны по коду
 * @param {string} countryCode - Код страны
 * @returns {Object|null} - Конфигурация страны или null
 */
export const getCountryConfig = (countryCode) => {
  return LICENSE_PLATE_COUNTRIES[countryCode] || null;
};

/**
 * Получает примеры валидных госномеров для указанной страны
 * @param {string} country - Код страны
 * @returns {Array<string>} - Массив примеров госномеров
 */
export const getLicensePlateExamples = (country = 'RUS') => {
  const countryConfig = LICENSE_PLATE_COUNTRIES[country];
  return countryConfig ? countryConfig.examples : [];
};

/**
 * Получает описание формата госномера для указанной страны
 * @param {string} country - Код страны
 * @returns {string} - Описание формата
 */
export const getLicensePlateFormatDescription = (country = 'RUS') => {
  const countryConfig = LICENSE_PLATE_COUNTRIES[country];
  if (!countryConfig) {
    return 'Неподдерживаемая страна';
  }
  return `Формат для ${countryConfig.name}: ${countryConfig.placeholder}`;
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
  const supportedLetters = ['A', 'B', 'E', 'K', 'M', 'H', 'O', 'P', 'C', 'T', 'Y', 'X'];
  
  return supportedLetters.includes(upperLetter);
};

/**
 * Получает список поддерживаемых букв для российских госномеров
 * 
 * @returns {Array<string>} - Массив поддерживаемых букв
 */
export const getSupportedLetters = () => {
  return ['A', 'B', 'E', 'K', 'M', 'H', 'O', 'P', 'C', 'T', 'Y', 'X'];
};

export default {
  validateAndNormalizeLicensePlate,
  isValidLicensePlate,
  normalizeLicensePlateForSearch,
  getLicensePlateVariants,
  formatLicensePlateForDisplay,
  getLicensePlateExamples,
  getLicensePlateFormatDescription,
  getSupportedCountries,
  getCountryConfig,
  isSupportedLetter,
  getSupportedLetters
};
