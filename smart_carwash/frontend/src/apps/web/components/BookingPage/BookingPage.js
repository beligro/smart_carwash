import React from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './BookingPage.module.css';
import ServiceSelector from '../ServiceSelector';

/**
 * Компонент BookingPage - страница записи на мойку
 * @param {Object} props - Свойства компонента
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Object} props.user - Данные пользователя
 * @param {Function} props.onCreateSession - Функция для создания сессии
 */
const BookingPage = ({ theme = 'light', user, onCreateSession }) => {
  const navigate = useNavigate();
  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  // Обработчик выбора услуги
  const handleServiceSelect = (serviceData) => {
    try {
      // Используем функцию создания сессии с платежом
      onCreateSession(serviceData);
    } catch (error) {
      alert('Ошибка при выборе услуги: ' + error.message);
    }
  };


  return (
    <div className={styles.container}>
      <div className={styles.content}>
        <ServiceSelector 
          onSelect={handleServiceSelect} 
          theme={theme} 
          user={user}
        />
      </div>
    </div>
  );
};

export default BookingPage;
