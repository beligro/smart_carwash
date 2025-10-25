import React, { useState, useEffect } from 'react';
import styles from './EmailInput.module.css';
import { Card } from '../../../../shared/components/UI';

/**
 * Компонент EmailInput - ввод email с валидацией
 * @param {Object} props - Свойства компонента
 * @param {string} props.value - Текущее значение email
 * @param {Function} props.onChange - Функция изменения значения
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {boolean} props.showRememberCheckbox - Показывать ли чекбокс "запомнить"
 * @param {boolean} props.rememberChecked - Состояние чекбокса "запомнить"
 * @param {Function} props.onRememberChange - Функция изменения чекбокса "запомнить"
 * @param {string} props.savedEmail - Сохраненный email пользователя
 */
const EmailInput = ({ 
  value, 
  onChange, 
  theme = 'light',
  showRememberCheckbox = false,
  rememberChecked = false,
  onRememberChange,
  savedEmail = ''
}) => {
  const [isValid, setIsValid] = useState(true);
  const [isFocused, setIsFocused] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  // Безопасное получение значения
  const safeValue = value || '';

  // Валидация email
  const validateEmail = (email) => {
    try {
      if (!email || typeof email !== 'string') {
        return { isValid: false, message: '' };
      }
      
      // Проверяем базовую структуру email
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) {
        return { isValid: false, message: 'Некорректный формат email' };
      }
      
      // Проверяем длину
      if (email.length > 254) {
        return { isValid: false, message: 'Email слишком длинный' };
      }
      
      return { isValid: true, message: '' };
    } catch (error) {
      console.error('Ошибка в validateEmail:', error);
      return { isValid: false, message: 'Ошибка валидации' };
    }
  };

  // Обработчик фокуса
  const handleFocus = () => {
    try {
      setIsFocused(true);
    } catch (error) {
      console.error('Ошибка в handleFocus:', error);
    }
  };

  // Обработчик потери фокуса
  const handleBlur = () => {
    try {
      setIsFocused(false);
      
      // Валидируем при потере фокуса
      if (safeValue) {
        const validation = validateEmail(safeValue);
        setIsValid(validation.isValid);
        setErrorMessage(validation.message);
      } else {
        setIsValid(true);
        setErrorMessage('');
      }
    } catch (error) {
      console.error('Ошибка в handleBlur:', error);
    }
  };

  // Обработчик ввода
  const handleInput = (e) => {
    try {
      const newValue = e.target.value;
      onChange(newValue);
      
      // Сбрасываем ошибку при вводе
      if (!isValid) {
        setIsValid(true);
        setErrorMessage('');
      }
    } catch (error) {
      console.error('Ошибка в handleInput:', error);
    }
  };

  // Определяем, нужно ли показывать чекбокс "запомнить"
  const shouldShowRememberCheckbox = showRememberCheckbox && 
    safeValue && 
    safeValue !== savedEmail && 
    isValid;

  return (
    <Card theme={theme} className={styles.container}>
      <div className={styles.inputGroup}>
        <label className={`${styles.label} ${themeClass}`}>
          Email для чека
        </label>
        <div className={`${styles.inputWrapper} ${!isValid ? styles.error : ''} ${isFocused ? styles.focused : ''}`}>
          <input
            type="email"
            value={safeValue}
            onChange={handleInput}
            onFocus={handleFocus}
            onBlur={handleBlur}
            placeholder="example@email.com"
            className={`${styles.input} ${themeClass}`}
            maxLength={254}
          />
          {!isValid && (
            <div className={styles.errorIcon}>⚠️</div>
          )}
        </div>
        {errorMessage && (
          <div className={styles.errorMessage}>{errorMessage}</div>
        )}
        <div className={styles.helpText}>
          Email будет использован для отправки чека об оплате
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
            Запомнить email
          </label>
          <div className={styles.rememberHelp}>
            Email будет автоматически подставляться при следующих записях
          </div>
        </div>
      )}
    </Card>
  );
};

export default EmailInput;
