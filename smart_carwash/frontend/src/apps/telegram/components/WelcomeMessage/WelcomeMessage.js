import React from 'react';
import WebApp from '@twa-dev/sdk';
import styles from './WelcomeMessage.module.css';
import { Card } from '../../../../shared/components/UI';

/**
 * Компонент WelcomeMessage - отображает приветственное сообщение
 * @param {Object} props - Свойства компонента
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 */
const WelcomeMessage = ({ theme = 'light' }) => {
  // Получаем имя пользователя из Telegram WebApp
  let firstName = 'Гость';
  
  try {
    firstName = WebApp.initDataUnsafe?.user?.first_name || 'Гость';
  } catch (err) {
    console.error('Ошибка получения данных пользователя:', err);
  }

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  return (
    <Card theme={theme}>
      <div className={`${styles.container} ${themeClass}`}>
        <h1 className={styles.title}>Привет, {firstName}! 👋</h1>
        <p className={styles.description}>
          Добро пожаловать в приложение умной автомойки. Здесь вы можете увидеть 
          информацию о доступных боксах и их текущем статусе.
        </p>
        <p className={styles.description}>
          Наша автомойка оборудована современными системами, которые обеспечивают 
          качественную и быструю мойку вашего автомобиля.
        </p>
      </div>
    </Card>
  );
};

export default WelcomeMessage;
