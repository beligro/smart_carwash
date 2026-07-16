import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
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
  const location = useLocation();
  const preselectServiceType = location.state?.preselectServiceType;
  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  // Обработчик выбора услуги. Возвращаем промис создания, чтобы ServiceSelector
  // мог дождаться завершения запроса и держать кнопку заблокированной.
  const handleServiceSelect = (serviceData) => {
    return onCreateSession(serviceData);
  };


  return (
    <div className={styles.container}>
      <div className={styles.content}>
        <ServiceSelector 
          onSelect={handleServiceSelect} 
          theme={theme} 
          user={user}
          initialServiceType={preselectServiceType}
        />
      </div>
    </div>
  );
};

export default BookingPage;
