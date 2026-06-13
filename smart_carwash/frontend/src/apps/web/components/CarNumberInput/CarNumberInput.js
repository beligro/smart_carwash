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
 * @param {boolean} props.noCarNumber - Состояние "нет номера"
 * @param {Function} props.onNoCarNumberChange - Функция изменения состояния "нет номера"
 * @param {boolean} props.showDisclaimer - Показывать ли дисклеймер с предупреждением
 * @param {boolean} props.showConfirmation - Показывать ли подтверждение номера
 * @param {boolean} props.confirmationChecked - Состояние чекбокса подтверждения
 * @param {Function} props.onConfirmationChange - Функция изменения чекбокса подтверждения
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
  savedCarNumber = '',
  noCarNumber = false,
  onNoCarNumberChange,
  showDisclaimer = false,
  showConfirmation = false,
  confirmationChecked = false,
  onConfirmationChange
}) => {
  const [isValid, setIsValid] = useState(true);
  const [errorMessage, setErrorMessage] = useState('');
  const [isFocused, setIsFocused] = useState(false);
  const [isComposing, setIsComposing] = useState(false);

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

  // Валидация номера машины для выбранной страны (только для внутренней логики)
  const validateCarNumber = (number, countryToValidate = country) => {
    try {
      // Используем новую утилиту для валидации и нормализации с указанием страны
      const validation = validateAndNormalizeLicensePlate(number, countryToValidate);
      
      // Всегда считаем валидным для отображения, ошибки показываем только через suggestion
      setIsValid(true);
      setErrorMessage('');
      return validation.isValid;
    } catch (error) {
      console.error('Ошибка валидации номера машины:', error);
      setIsValid(true);
      setErrorMessage('');
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

  // Обработчик изменения чекбокса "Нет номера"
  const handleNoCarNumberChange = (e) => {
    try {
      const checked = e.target.checked;
      onNoCarNumberChange(checked);
      
      // Если включаем "нет номера", очищаем номер
      if (checked) {
        onChange('');
      }
    } catch (error) {
      console.error('Ошибка в handleNoCarNumberChange:', error);
    }
  };

  // Обработчик изменения чекбокса подтверждения номера
  const handleConfirmationChange = (e) => {
    try {
      const checked = e.target.checked;
      onConfirmationChange(checked);
    } catch (error) {
      console.error('Ошибка в handleConfirmationChange:', error);
    }
  };

  // Обработчик потери фокуса
  const handleBlur = () => {
    try {
      setIsFocused(false);
      // На blur нормализуем и валидируем
      const normalized = formatCarNumber((safeValue || '').toUpperCase());
      if (normalized !== safeValue) {
        onChange(normalized);
      }
      validateCarNumber(normalized);
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
      const currentValue = e.target.value;
      if (isComposing) {
        onChange(currentValue);
        return;
      }
      // В обычном режиме не навязываем формат при каждом вводе,
      // просто передаем текущее значение; нормализуем на blur/compositionend
      onChange(currentValue);
    } catch (error) {
      console.error('Ошибка в handleInput:', error);
    }
  };

  // Композиция ввода для совместимости со старыми клавиатурами Android
  const handleCompositionStart = () => {
    try {
      setIsComposing(true);
    } catch (error) {
      console.error('Ошибка в handleCompositionStart:', error);
    }
  };

  const handleCompositionEnd = (e) => {
    try {
      setIsComposing(false);
      // Завершение ввода: безопасная нормализация
      const normalized = formatCarNumber((e.target.value || '').toUpperCase());
      onChange(normalized);
      setTimeout(() => {
        validateCarNumber(normalized);
      }, 0);
    } catch (error) {
      console.error('Ошибка в handleCompositionEnd:', error);
    }
  };

  // Определяем, нужно ли показывать чекбокс "запомнить"
  const shouldShowRememberCheckbox = showRememberCheckbox && 
    safeValue && 
    safeValue !== savedCarNumber && 
    isValid;

  return (
    <Card theme={theme} className={styles.container}>
      {/* Чекбокс "Нет номера" */}
      <div className={styles.noCarNumberContainer}>
        <div className={styles.optionRow}>
          <label className={`${styles.optionLabel} ${themeClass}`}>
            <input
              type="checkbox"
              checked={noCarNumber}
              onChange={handleNoCarNumberChange}
              className={styles.checkbox}
            />
            <span className={styles.checkmark}></span>
            У меня нет номера на автомобиле
          </label>
        </div>
      </div>

      {/* Сообщение для автомобилей без номера */}
      {noCarNumber && (
        <div className={`${styles.noCarNumberMessage} ${themeClass}`}>
          <div className={styles.messageIcon}>⚠️</div>
          <div className={styles.messageText}>
            Для автомобилей без номера необходимо записаться и оплатить мойку у кассира-администратора
          </div>
        </div>
      )}

      {/* Поля ввода номера - показываем только если НЕ выбрано "нет номера" */}
      {!noCarNumber && (
        <>
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
            <div className={`${styles.inputWrapper} ${isFocused ? styles.focused : ''}`}>
              <input
                type="text"
                value={safeValue}
                onChange={handleInput}
                onCompositionStart={handleCompositionStart}
                onCompositionEnd={handleCompositionEnd}
                onFocus={handleFocus}
                onBlur={handleBlur}
                placeholder={countryConfig?.placeholder || "А123ВК456"}
                className={`${styles.input} ${themeClass}`}
                maxLength={12}
                autoCapitalize="characters"
                autoCorrect="off"
                spellCheck={false}
                inputMode="text"
              />
            </div>
            <div className={styles.helpText}>
              💡 Проверьте формат номера. Пример: А123ВС77
            </div>
          </div>

          {/* Дисклеймер с предупреждением */}
          {showDisclaimer && (
            <div className={`${styles.disclaimerContainer} ${themeClass}`}>
              <div className={styles.disclaimerIcon}>⚠️</div>
              <div className={styles.disclaimerContent}>
                <div className={styles.disclaimerTitle}>Внимание! При вводе некорректного номера возможны:</div>
                <ul className={styles.disclaimerList}>
                  <li>Преждевременное отключение оборудования</li>
                  <li>Некорректное время работы поста</li>
                  <li>Автоматическое завершение сессии</li>
                </ul>
              </div>
            </div>
          )}

          {/* Подтверждение номера */}
          {showConfirmation && (
            <div className={styles.confirmationContainer}>
              <div className={styles.optionRow}>
                <label className={`${styles.optionLabel} ${themeClass}`}>
                  <input
                    type="checkbox"
                    checked={confirmationChecked}
                    onChange={handleConfirmationChange}
                    className={styles.checkbox}
                  />
                  <span className={styles.checkmark}></span>
                  Подтверждаю, что ввёл корректный номер моего автомобиля и понимаю возможные последствия
                </label>
              </div>
            </div>
          )}

          {shouldShowRememberCheckbox && (
            <div className={styles.rememberContainer}>
              <div className={styles.optionRow}>
                <label className={`${styles.optionLabel} ${themeClass}`}>
                  <input
                    type="checkbox"
                    checked={rememberChecked}
                    onChange={(e) => onRememberChange(e.target.checked)}
                    className={styles.checkbox}
                  />
                  <span className={styles.checkmark}></span>
                  Запомнить номер машины
                </label>
              </div>
              <div className={styles.rememberHelp}>
                Номер будет автоматически подставляться при следующих записях
              </div>
            </div>
          )}
        </>
      )}
    </Card>
  );
};

export default CarNumberInput; 