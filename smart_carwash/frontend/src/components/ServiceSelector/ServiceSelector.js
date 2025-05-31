import React, { useState } from 'react';
import styles from './ServiceSelector.module.css';
import { Card, Button } from '../UI';

/**
 * Компонент ServiceSelector - позволяет выбрать тип услуги и дополнительные опции
 * @param {Object} props - Свойства компонента
 * @param {Function} props.onSelect - Функция, вызываемая при выборе услуги
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 */
const ServiceSelector = ({ onSelect, theme = 'light' }) => {
  const [selectedService, setSelectedService] = useState(null);
  const [withChemistry, setWithChemistry] = useState(false);
  
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  // Типы услуг
  const serviceTypes = [
    { id: 'wash', name: 'Мойка', description: 'Стандартная мойка автомобиля', hasChemistry: true },
    { id: 'air_dry', name: 'Обдув воздухом', description: 'Сушка автомобиля воздухом', hasChemistry: false },
    { id: 'vacuum', name: 'Пылесос', description: 'Уборка салона пылесосом', hasChemistry: false }
  ];
  
  // Обработчик выбора услуги
  const handleServiceSelect = (serviceType) => {
    setSelectedService(serviceType);
    // Если выбрана услуга без химии, сбрасываем флаг
    if (!serviceType.hasChemistry) {
      setWithChemistry(false);
    }
  };
  
  // Обработчик переключения опции химии
  const handleChemistryToggle = () => {
    setWithChemistry(!withChemistry);
  };
  
  // Обработчик подтверждения выбора
  const handleConfirm = () => {
    if (selectedService) {
      onSelect({
        serviceType: selectedService.id,
        withChemistry: selectedService.hasChemistry ? withChemistry : false
      });
    }
  };
  
  return (
    <div className={styles.container}>
      <h2 className={`${styles.title} ${themeClass}`}>Выберите услугу</h2>
      
      <div className={styles.serviceGrid}>
        {serviceTypes.map((service) => (
          <Card 
            key={service.id} 
            theme={theme} 
            className={`${styles.serviceCard} ${selectedService?.id === service.id ? styles.selected : ''}`}
            onClick={() => handleServiceSelect(service)}
          >
            <h3 className={`${styles.serviceName} ${themeClass}`}>{service.name}</h3>
            <p className={`${styles.serviceDescription} ${themeClass}`}>{service.description}</p>
            {selectedService?.id === service.id && (
              <div className={styles.selectedIndicator}></div>
            )}
          </Card>
        ))}
      </div>
      
      {selectedService?.hasChemistry && (
        <div className={styles.optionsContainer}>
          <Card theme={theme} className={styles.optionCard}>
            <div className={styles.optionRow}>
              <label className={`${styles.optionLabel} ${themeClass}`}>
                <input 
                  type="checkbox" 
                  checked={withChemistry} 
                  onChange={handleChemistryToggle}
                  className={styles.checkbox}
                />
                <span className={styles.checkmark}></span>
                Использовать химию
              </label>
            </div>
            <p className={`${styles.optionDescription} ${themeClass}`}>
              Химия помогает лучше очистить поверхность автомобиля от грязи и жира
            </p>
          </Card>
        </div>
      )}
      
      <div className={styles.buttonContainer}>
        <Button 
          theme={theme} 
          onClick={handleConfirm}
          disabled={!selectedService}
          className={styles.confirmButton}
        >
          Подтвердить выбор
        </Button>
      </div>
    </div>
  );
};

export default ServiceSelector;
