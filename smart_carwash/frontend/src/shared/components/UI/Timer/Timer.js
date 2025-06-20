import React from 'react';
import styles from './Timer.module.css';

/**
 * Компонент Timer - отображает таймер обратного отсчета
 * @param {Object} props - Свойства компонента
 * @param {number} props.seconds - Количество секунд для отображения
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {string} props.className - Дополнительные CSS классы
 * @param {Object} props.style - Дополнительные inline стили
 */
const Timer = ({ 
  seconds, 
  theme = 'light', 
  className = '', 
  style = {},
  ...props 
}) => {
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  // Определяем класс в зависимости от оставшегося времени
  let timeClass = '';
  if (seconds <= 60) {
    timeClass = styles.danger; // Красный для последней минуты
  } else if (seconds <= 120) {
    timeClass = styles.warning; // Оранжевый для 1-2 минут
  }
  
  // Форматируем время в формат MM:SS
  const formattedTime = formatTime(seconds);
  
  return (
    <div 
      className={`${styles.timer} ${themeClass} ${timeClass} ${className}`}
      style={style}
      {...props}
    >
      {formattedTime}
    </div>
  );
};

/**
 * Функция для форматирования времени в формат MM:SS
 * @param {number} seconds - Количество секунд
 * @returns {string} - Отформатированное время
 */
const formatTime = (seconds) => {
  if (seconds === null || seconds === undefined) return '00:00';
  
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  
  return `${minutes}:${remainingSeconds < 10 ? '0' : ''}${remainingSeconds}`;
};

export default Timer;
