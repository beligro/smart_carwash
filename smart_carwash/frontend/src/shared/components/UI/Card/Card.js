import React from 'react';
import styles from './Card.module.css';

/**
 * Компонент Card - универсальный компонент для отображения карточек
 * @param {Object} props - Свойства компонента
 * @param {React.ReactNode} props.children - Дочерние элементы
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {string} props.className - Дополнительные CSS классы
 * @param {Object} props.style - Дополнительные inline стили
 */
const Card = ({ children, theme = 'light', className = '', style = {}, ...props }) => {
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  return (
    <div 
      className={`${styles.card} ${themeClass} ${className}`}
      style={style}
      {...props}
    >
      {children}
    </div>
  );
};

export default Card;
