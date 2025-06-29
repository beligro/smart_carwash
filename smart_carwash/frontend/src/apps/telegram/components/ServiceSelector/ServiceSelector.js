import React, { useState, useEffect } from 'react';
import styles from './ServiceSelector.module.css';
import { Card, Button } from '../../../../shared/components/UI';
import CarNumberInput from '../CarNumberInput';
import ApiService from '../../../../shared/services/ApiService';

/**
 * Компонент ServiceSelector - позволяет выбрать тип услуги и дополнительные опции
 * @param {Object} props - Свойства компонента
 * @param {Function} props.onSelect - Функция, вызываемая при выборе услуги
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Object} props.user - Данные пользователя (для получения сохраненного номера)
 */
const ServiceSelector = ({ onSelect, theme = 'light', user }) => {
  const [selectedService, setSelectedService] = useState(null);
  const [withChemistry, setWithChemistry] = useState(false);
  const [rentalTimes, setRentalTimes] = useState([]);
  const [selectedRentalTime, setSelectedRentalTime] = useState(null);
  const [loading, setLoading] = useState(false);
  const [carNumber, setCarNumber] = useState('');
  const [rememberCarNumber, setRememberCarNumber] = useState(false);
  const [savingCarNumber, setSavingCarNumber] = useState(false);
  
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  // Инициализация номера машины из данных пользователя
  useEffect(() => {
    if (user && user.car_number) {
      setCarNumber(user.car_number);
    }
  }, [user]);
  
  // Типы услуг
  const serviceTypes = [
    { id: 'wash', name: 'Мойка', description: 'Стандартная мойка автомобиля', hasChemistry: true },
    { id: 'air_dry', name: 'Обдув воздухом', description: 'Сушка автомобиля воздухом', hasChemistry: false },
    { id: 'vacuum', name: 'Пылесос', description: 'Уборка салона пылесосом', hasChemistry: false }
  ];
  
  // Загрузка доступного времени аренды при выборе услуги
  useEffect(() => {
    if (selectedService) {
      setLoading(true);
      ApiService.getAvailableRentalTimes(selectedService.id)
        .then(data => {
          setRentalTimes(data.available_times || [5]);
          setSelectedRentalTime(data.available_times && data.available_times.length > 0 ? data.available_times[0] : 5);
        })
        .catch(error => {
          console.error('Ошибка при загрузке времени аренды:', error);
          setRentalTimes([5]);
          setSelectedRentalTime(5);
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setRentalTimes([]);
      setSelectedRentalTime(null);
    }
  }, [selectedService]);

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
  
  // Обработчик выбора времени аренды
  const handleRentalTimeSelect = (time) => {
    setSelectedRentalTime(time);
  };

  // Обработчик изменения номера машины
  const handleCarNumberChange = (value) => {
    setCarNumber(value);
  };

  // Обработчик изменения чекбокса "запомнить номер"
  const handleRememberCarNumberChange = (checked) => {
    setRememberCarNumber(checked);
  };

  // Сохранение номера машины пользователя
  const saveCarNumber = async () => {
    if (!user || !carNumber || !rememberCarNumber) {
      return;
    }

    setSavingCarNumber(true);
    try {
      await ApiService.updateCarNumber(user.id, carNumber);
      console.log('Номер машины сохранен');
    } catch (error) {
      console.error('Ошибка при сохранении номера машины:', error);
    } finally {
      setSavingCarNumber(false);
    }
  };

  // Обработчик подтверждения выбора
  const handleConfirm = async () => {
    if (selectedService && selectedRentalTime && carNumber) {
      // Если пользователь хочет запомнить номер, сохраняем его
      if (rememberCarNumber) {
        await saveCarNumber();
      }

      onSelect({
        serviceType: selectedService.id,
        withChemistry: selectedService.hasChemistry ? withChemistry : false,
        rentalTimeMinutes: selectedRentalTime,
        carNumber: carNumber
      });
    }
  };

  // Проверяем, можно ли подтвердить выбор
  const canConfirm = selectedService && 
    selectedRentalTime && 
    carNumber && 
    carNumber.length === 9 && 
    /^[АВЕКМНОРСТУХ]\d{3}[АВЕКМНОРСТУХ]{2}\d{2,3}$/.test(carNumber);

  // Определяем, нужно ли показывать чекбокс "запомнить номер"
  const showRememberCheckbox = user && 
    carNumber && 
    carNumber !== user.car_number && 
    /^[АВЕКМНОРСТУХ]\d{3}[АВЕКМНОРСТУХ]{2}\d{2,3}$/.test(carNumber);
  
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
      
      {selectedService && (
        <div className={styles.optionsContainer}>
          {/* Ввод номера машины */}
          <CarNumberInput
            value={carNumber}
            onChange={handleCarNumberChange}
            theme={theme}
            showRememberCheckbox={showRememberCheckbox}
            rememberChecked={rememberCarNumber}
            onRememberChange={handleRememberCarNumberChange}
            savedCarNumber={user?.car_number || ''}
          />

          {selectedService.hasChemistry && (
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
          )}
          
          <Card theme={theme} className={styles.optionCard}>
            <h3 className={`${styles.optionTitle} ${themeClass}`}>Выберите время аренды</h3>
            {loading ? (
              <p className={`${styles.loadingText} ${themeClass}`}>Загрузка доступного времени...</p>
            ) : (
              <div className={styles.rentalTimeGrid}>
                {rentalTimes.map((time) => (
                  <div 
                    key={time} 
                    className={`${styles.rentalTimeItem} ${selectedRentalTime === time ? styles.selectedTime : ''}`}
                    onClick={() => handleRentalTimeSelect(time)}
                  >
                    <span className={`${styles.rentalTimeValue} ${themeClass}`}>{time}</span>
                    <span className={`${styles.rentalTimeUnit} ${themeClass}`}>мин</span>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </div>
      )}
      
      <div className={styles.buttonContainer}>
        <Button 
          theme={theme} 
          onClick={handleConfirm}
          disabled={!canConfirm || loading || savingCarNumber}
          className={styles.confirmButton}
        >
          {savingCarNumber ? 'Сохранение...' : 'Подтвердить выбор'}
        </Button>
      </div>
    </div>
  );
};

export default ServiceSelector;
