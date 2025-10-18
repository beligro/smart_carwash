import React, { useState, useEffect } from 'react';
import styles from './PriceCalculator.module.css';
import ApiService from '../../../../shared/services/ApiService';

/**
 * Компонент PriceCalculator - рассчитывает и показывает цену услуги
 * @param {Object} props - Свойства компонента
 * @param {string} props.serviceType - Тип услуги
 * @param {boolean} props.withChemistry - Использование химии
 * @param {number} props.chemistryTimeMinutes - Время химии в минутах
 * @param {number} props.rentalTimeMinutes - Время аренды в минутах
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 */
const PriceCalculator = ({ serviceType, withChemistry, chemistryTimeMinutes, rentalTimeMinutes, theme = 'light' }) => {
  const [price, setPrice] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  // Загружаем цену при изменении параметров
  useEffect(() => {
    if (serviceType && rentalTimeMinutes) {
      calculatePrice();
    }
  }, [serviceType, withChemistry, chemistryTimeMinutes, rentalTimeMinutes]);

  const calculatePrice = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await ApiService.calculatePrice({
        service_type: serviceType,
        with_chemistry: withChemistry,
        chemistry_time_minutes: chemistryTimeMinutes || 0,
        rental_time_minutes: rentalTimeMinutes
      });
      
      setPrice(response);
    } catch (err) {
      console.error('Ошибка расчета цены:', err);
      setError('Не удалось рассчитать цену');
    } finally {
      setLoading(false);
    }
  };

  // Форматирование цены в рубли
  const formatPrice = (priceInKopecks) => {
    if (!priceInKopecks) return '0 ₽';
    return `${(priceInKopecks / 100).toFixed(0)} ₽`;
  };

  // Получение названия услуги
  const getServiceName = (type) => {
    const services = {
      'wash': 'Мойка',
      'air_dry': 'Обдув',
      'vacuum': 'Пылесос'
    };
    return services[type] || type;
  };

  if (loading) {
    return (
      <div className={`${styles.priceCalculator} ${themeClass}`}>
        <div className={styles.loading}>Расчет цены...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`${styles.priceCalculator} ${themeClass}`}>
        <div className={styles.error}>{error}</div>
      </div>
    );
  }

  if (!price) {
    return null;
  }

  return (
    <div className={`${styles.priceCalculator} ${themeClass}`}>
      <div className={styles.priceInfo}>
        <div className={styles.serviceName}>
          {getServiceName(serviceType)}
          {withChemistry && serviceType === 'wash' && (
            <span className={styles.chemistryBadge}>+ химия</span>
          )}
        </div>
        <div className={styles.duration}>
          {rentalTimeMinutes} мин
        </div>
        <div className={styles.price}>
          {formatPrice(price.price)}
        </div>
      </div>
      
      {price.breakdown && (
        <div className={styles.breakdown}>
          <div className={styles.breakdownItem}>
            <span>Базовая цена:</span>
            <span>{formatPrice(price.breakdown.base_price)}</span>
          </div>
          {price.breakdown.chemistry_price > 0 && (
            <div className={styles.breakdownItem}>
              <span>Химия ({chemistryTimeMinutes} мин):</span>
              <span>{formatPrice(price.breakdown.chemistry_price)}</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default PriceCalculator; 