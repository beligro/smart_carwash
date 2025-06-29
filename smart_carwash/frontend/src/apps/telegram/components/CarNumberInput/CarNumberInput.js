import React, { useState, useEffect } from 'react';
import styles from './CarNumberInput.module.css';
import { Card } from '../../../../shared/components/UI';

/**
 * Компонент CarNumberInput - ввод номера машины с валидацией
 * @param {Object} props - Свойства компонента
 * @param {string} props.value - Текущее значение номера
 * @param {Function} props.onChange - Функция изменения значения
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {boolean} props.showRememberCheckbox - Показывать ли чекбокс "запомнить"
 * @param {boolean} props.rememberChecked - Состояние чекбокса "запомнить"
 * @param {Function} props.onRememberChange - Функция изменения чекбокса "запомнить"
 * @param {string} props.savedCarNumber - Сохраненный номер машины пользователя
 */
const CarNumberInput = ({ 
  value, 
  onChange, 
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

  // Валидация российского номера машины (формат: О111ОО799)
  const validateCarNumber = (number) => {
    if (!number) {
      setIsValid(false);
      setErrorMessage('Введите номер машины');
      return false;
    }

    // Регулярное выражение для российского номера
    const carNumberRegex = /^[АВЕКМНОРСТУХ]\d{3}[АВЕКМНОРСТУХ]{2}\d{2,3}$/;
    
    if (!carNumberRegex.test(number)) {
      setIsValid(false);
      setErrorMessage('Неверный формат номера. Используйте формат: О111ОО799');
      return false;
    }

    setIsValid(true);
    setErrorMessage('');
    return true;
  };

  // Обработчик изменения значения
  const handleChange = (e) => {
    const newValue = e.target.value.toUpperCase();
    onChange(newValue);
    validateCarNumber(newValue);
  };

  // Обработчик потери фокуса
  const handleBlur = () => {
    setIsFocused(false);
    validateCarNumber(value);
  };

  // Обработчик получения фокуса
  const handleFocus = () => {
    setIsFocused(true);
  };

  // Автоматическое форматирование при вводе
  const formatCarNumber = (input) => {
    // Убираем все пробелы и дефисы
    let formatted = input.replace(/[\s-]/g, '');
    
    // Ограничиваем длину
    if (formatted.length > 9) {
      formatted = formatted.substring(0, 9);
    }
    
    return formatted;
  };

  // Обработчик ввода с форматированием
  const handleInput = (e) => {
    const formatted = formatCarNumber(e.target.value.toUpperCase());
    onChange(formatted);
  };

  // Определяем, нужно ли показывать чекбокс "запомнить"
  const shouldShowRememberCheckbox = showRememberCheckbox && 
    value && 
    value !== savedCarNumber && 
    validateCarNumber(value);

  return (
    <Card theme={theme} className={styles.container}>
      <div className={styles.inputGroup}>
        <label className={`${styles.label} ${themeClass}`}>
          Номер машины
        </label>
        <div className={`${styles.inputWrapper} ${!isValid ? styles.error : ''} ${isFocused ? styles.focused : ''}`}>
          <input
            type="text"
            value={value}
            onChange={handleInput}
            onFocus={handleFocus}
            onBlur={handleBlur}
            placeholder="О111ОО799"
            className={`${styles.input} ${themeClass}`}
            maxLength={9}
          />
          {!isValid && (
            <div className={styles.errorIcon}>⚠️</div>
          )}
        </div>
        {errorMessage && (
          <div className={styles.errorMessage}>{errorMessage}</div>
        )}
        <div className={styles.helpText}>
          Введите номер в формате: О111ОО799
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