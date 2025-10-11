import React from 'react';
import styles from './Header.module.css';

/**
 * Компонент Header - отображает заголовок приложения
 * @param {Object} props - Свойства компонента
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Function} props.onBack - Функция для обработки нажатия на стрелку назад (опционально)
 */
const Header = ({ theme = 'light', onBack }) => {
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  return (
    <header className={`${styles.header} ${themeClass}`}>
      {onBack && (
        <button 
          onClick={onBack} 
          className={`${styles.backButton} ${themeClass}`}
          style={{
            background: 'none',
            border: 'none',
            fontSize: '24px',
            cursor: 'pointer',
            padding: '8px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            position: 'absolute',
            left: '16px',
            top: '50%',
            transform: 'translateY(-50%)'
          }}
        >
          ←
        </button>
      )}
      <div className={styles.logo}>
        <span className={styles.logoIcon}>🚿</span>
        Автомойка H2O
      </div>
    </header>
  );
};

export default Header;
