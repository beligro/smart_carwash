import React from 'react';
import styles from './StatusBadge.module.css';

/**
 * Компонент StatusBadge - отображает статус в виде бейджа
 * @param {Object} props - Свойства компонента
 * @param {string} props.status - Статус ('created', 'in_queue', 'payment_failed', 'assigned', 'active', 'complete', 'canceled', 'expired', 'free', 'reserved', 'busy', 'maintenance')
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {string} props.text - Текст для отображения (если не указан, будет использован статус)
 * @param {string} props.className - Дополнительные CSS классы
 * @param {Object} props.style - Дополнительные inline стили
 */
const StatusBadge = ({ 
  status, 
  theme = 'light', 
  text, 
  className = '', 
  style = {},
  ...props 
}) => {
  const themeClass = theme === 'dark' ? styles.dark : '';
  const statusClass = styles[status] || '';
  
  // Определяем текст для отображения
  const displayText = text || getStatusText(status);
  
  return (
    <div 
      className={`${styles.badge} ${themeClass} ${statusClass} ${className}`}
      style={style}
      {...props}
    >
      {displayText}
    </div>
  );
};

/**
 * Функция для получения текста статуса
 * @param {string} status - Статус
 * @returns {string} - Текст статуса
 */
const getStatusText = (status) => {
  switch (status) {
    // Статусы сессий
    case 'created':
      return 'Создана';
    case 'in_queue':
      return 'В очереди';
    case 'payment_failed':
      return 'Ошибка оплаты';
    case 'assigned':
      return 'Назначена';
    case 'active':
      return 'Активна';
    case 'complete':
      return 'Завершена';
    case 'canceled':
      return 'Отменена';
    case 'expired':
      return 'Истекла';
      
    // Статусы боксов
    case 'free':
      return 'Свободен';
    case 'reserved':
      return 'Зарезервирован';
    case 'busy':
      return 'Занят';
    case 'maintenance':
      return 'На обслуживании';
      
    default:
      return 'Неизвестно';
  }
};

export default StatusBadge;
