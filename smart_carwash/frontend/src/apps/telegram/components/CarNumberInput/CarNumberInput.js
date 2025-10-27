import React, { useState, useEffect } from 'react';
import styles from './CarNumberInput.module.css';
import { Card } from '../../../../shared/components/UI';
import { 
  validateAndNormalizeLicensePlate, 
  formatLicensePlateForDisplay, 
  getLicensePlateExamples, 
  getLicensePlateFormatDescription,
  getSupportedCountries,
  getCountryConfig
} from '../../../../shared/utils/licensePlateUtils';

/**
 * Компонент CarNumberInput - ввод номера машины с валидацией и выбором страны
 * @param {Object} props - Свойства компонента
 * @param {string} props.value - Текущее значение номера
 * @param {Function} props.onChange - Функция изменения значения
 * @param {string} props.country - Текущая выбранная страна
 * @param {Function} props.onCountryChange - Функция изменения страны
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {boolean} props.showRememberCheckbox - Показывать ли чекбокс "запомнить"
 * @param {boolean} props.rememberChecked - Состояние чекбокса "запомнить"
 * @param {Function} props.onRememberChange - Функция изменения чекбокса "запомнить"
 * @param {string} props.savedCarNumber - Сохраненный номер машины пользователя
 */
const CarNumberInput = ({ 
  value, 
  onChange, 
  country = 'RUS',
  onCountryChange,
  theme = 'light',
  showRememberCheckbox = false,
  rememberChecked = false,
  onRememberChange,
  savedCarNumber = ''
}) => {
  const [isValid, setIsValid] = useState(true);
  const [errorMessage, setErrorMessage] = useState('');
  const [isFocused, setIsFocused] = useState(false);

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  // Обеспечиваем безопасность value
  const safeValue = value || '';

  // Получаем конфигурацию текущей страны
  const countryConfig = getCountryConfig(country);

  // Перевалидируем номер при изменении страны
  useEffect(() => {
    if (safeValue) {
      validateCarNumber(safeValue, country);
    }
  }, [country]);

  // Валидация номера машины для выбранной страны
  const validateCarNumber = (number, countryToValidate = country) => {
    try {
      // Используем новую утилиту для валидации и нормализации с указанием страны
      const validation = validateAndNormalizeLicensePlate(number, countryToValidate);
      
      if (!validation.isValid) {
        setIsValid(false);
        setErrorMessage(validation.error);
        return false;
      }

      setIsValid(true);
      setErrorMessage('');
      return true;
    } catch (error) {
      console.error('Ошибка валидации номера машины:', error);
      setIsValid(false);
      setErrorMessage('Ошибка валидации');
      return false;
    }
  };

  // Обработчик изменения значения
  const handleChange = (e) => {
    try {
      const inputValue = e.target.value.toUpperCase();
      
      // Нормализуем номер при вводе для выбранной страны
      const validation = validateAndNormalizeLicensePlate(inputValue, country);
      
      // Если номер валидный, используем нормализованную версию
      // Если не валидный, используем исходное значение для продолжения ввода
      const valueToSet = validation.isValid ? validation.normalized : inputValue;
      
      onChange(valueToSet);
      
      // Валидируем с задержкой для лучшего UX
      setTimeout(() => {
        validateCarNumber(valueToSet);
      }, 300);
    } catch (error) {
      console.error('Ошибка в handleChange:', error);
    }
  };

  // Обработчик изменения страны
  const handleCountryChange = (e) => {
    try {
      const newCountry = e.target.value;
      onCountryChange(newCountry);
      // Валидация произойдет автоматически через useEffect
    } catch (error) {
      console.error('Ошибка в handleCountryChange:', error);
    }
  };

  // Обработчик потери фокуса
  const handleBlur = () => {
    try {
      setIsFocused(false);
      validateCarNumber(safeValue);
    } catch (error) {
      console.error('Ошибка в handleBlur:', error);
    }
  };

  // Обработчик получения фокуса
  const handleFocus = () => {
    try {
      setIsFocused(true);
    } catch (error) {
      console.error('Ошибка в handleFocus:', error);
    }
  };

  // Автоматическое форматирование при вводе
  const formatCarNumber = (input) => {
    try {
      // Проверяем, что input - это строка
      if (typeof input !== 'string') {
        return '';
      }
      
      // Убираем все пробелы и дефисы
      let formatted = input.replace(/[\s-]/g, '');
      
      // Ограничиваем длину (увеличиваем до 12 символов)
      if (formatted.length > 12) {
        formatted = formatted.substring(0, 12);
      }
      
      return formatted;
    } catch (error) {
      console.error('Ошибка в formatCarNumber:', error);
      return '';
    }
  };

  // Обработчик ввода с форматированием
  const handleInput = (e) => {
    try {
      const formatted = formatCarNumber(e.target.value.toUpperCase());
      onChange(formatted);
    } catch (error) {
      console.error('Ошибка в handleInput:', error);
    }
  };

  // Определяем, нужно ли показывать чекбокс "запомнить"
  const shouldShowRememberCheckbox = showRememberCheckbox && 
    safeValue && 
    safeValue !== savedCarNumber && 
    isValid;

  return (
    <Card theme={theme} className={styles.container}>
      <div className={styles.inputGroup}>
        <label className={`${styles.label} ${themeClass}`}>
          Страна гос номера
        </label>
        <select
          value={country}
          onChange={handleCountryChange}
          className={`${styles.countrySelect} ${themeClass}`}
        >
          {getSupportedCountries().map(countryOption => (
            <option key={countryOption.code} value={countryOption.code}>
              {countryOption.name}
            </option>
          ))}
        </select>
      </div>

      <div className={styles.inputGroup}>
        <label className={`${styles.label} ${themeClass}`}>
          Номер машины
        </label>
        <div className={`${styles.inputWrapper} ${!isValid ? styles.error : ''} ${isFocused ? styles.focused : ''}`}>
          <input
            type="text"
            value={safeValue}
            onChange={handleInput}
            onFocus={handleFocus}
            onBlur={handleBlur}
            placeholder={countryConfig?.placeholder || "А123ВК456"}
            className={`${styles.input} ${themeClass}`}
            maxLength={12}
          />
          {!isValid && safeValue && (
            <div className={styles.errorIcon}>⚠️</div>
          )}
          {isValid && safeValue && (
            <div className={styles.successIcon}>✅</div>
          )}
        </div>
        {errorMessage && (
          <div className={styles.errorMessage}>{errorMessage}</div>
        )}
        <div className={styles.helpText}>
          {getLicensePlateFormatDescription(country)}
        </div>
        <div className={styles.examplesText}>
          Примеры: {getLicensePlateExamples(country).join(', ')}
        </div>
      </div>

      {shouldShowRememberCheckbox && (
        <div className={styles.rememberContainer}>
          <label className={`${styles.rememberLabel} ${themeClass}`}>
            <input
              type="checkbox"
              checked={rememberChecked}
              onChange={(e) => onRememberChange(e.target.checked)}
              className={styles.checkbox}
            />
            <span className={styles.checkmark}></span>
            Запомнить номер машины
          </label>
          <div className={styles.rememberHelp}>
            Номер будет автоматически подставляться при следующих записях
          </div>
        </div>
      )}
    </Card>
  );
};

export default CarNumberInput; 