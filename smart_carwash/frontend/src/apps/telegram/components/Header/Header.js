import React from 'react';
import styles from './Header.module.css';

/**
 * Компонент Header - отображает заголовок приложения
 * @param {Object} props - Свойства компонента
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 */
const Header = ({ theme = 'light' }) => {
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  return (
    <header className={`${styles.header} ${themeClass}`}>
      <div className={styles.logo}>
        <span className={styles.logoIcon}>🚿</span>
        Умная автомойка
      </div>
    </header>
  );
};

export default Header;
