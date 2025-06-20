import React from 'react';
import styles from './Button.module.css';

/**
 * Компонент Button - универсальный компонент кнопки
 * @param {Object} props - Свойства компонента
 * @param {React.ReactNode} props.children - Дочерние элементы
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {string} props.variant - Вариант кнопки ('primary', 'secondary', 'danger', 'success')
 * @param {boolean} props.disabled - Флаг отключения кнопки
 * @param {boolean} props.loading - Флаг загрузки
 * @param {function} props.onClick - Обработчик клика
 * @param {string} props.className - Дополнительные CSS классы
 * @param {Object} props.style - Дополнительные inline стили
 */
const Button = ({ 
  children, 
  theme = 'light', 
  variant = 'primary', 
  disabled = false, 
  loading = false,
  onClick,
  className = '',
  style = {},
  ...props 
}) => {
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  const variantClass = styles[variant] || styles.primary;
  const disabledClass = disabled || loading ? styles.disabled : '';
  
  return (
    <button 
      className={`${styles.button} ${themeClass} ${variantClass} ${disabledClass} ${className}`}
      onClick={onClick}
      disabled={disabled || loading}
      style={style}
      {...props}
    >
      {loading && <span className={styles.spinner}></span>}
      {children}
    </button>
  );
};

export default Button;
