import React, { useState, useEffect } from 'react';
import styles from './ServiceSelector.module.css';
import { Card, Button } from '../../../../shared/components/UI';
import CarNumberInput from '../CarNumberInput';
import ApiService from '../../../../shared/services/ApiService';
import PaymentModal from '../PaymentModal';
import PaymentService, { formatAmount } from '../../../../shared/services/PaymentService';

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
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [selectedServiceData, setSelectedServiceData] = useState(null);
  const [priceData, setPriceData] = useState(null);
  const [priceLoading, setPriceLoading] = useState(false);
  
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  // Инициализация номера машины из данных пользователя
  useEffect(() => {
    try {
      if (user && user.car_number) {
        setCarNumber(user.car_number);
      }
    } catch (error) {
      console.error('Ошибка при инициализации номера машины:', error);
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
    try {
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
    } catch (error) {
      console.error('Ошибка в useEffect для загрузки времени аренды:', error);
      setRentalTimes([5]);
      setSelectedRentalTime(5);
      setLoading(false);
    }
  }, [selectedService]);

  // Расчет цены при изменении параметров
  useEffect(() => {
    const calculatePrice = async () => {
      if (!selectedService || !selectedRentalTime) {
        setPriceData(null);
        return;
      }

      setPriceLoading(true);
      try {
        const priceResponse = await PaymentService.calculatePrice(
          selectedService.id,
          selectedRentalTime,
          selectedService.hasChemistry ? withChemistry : false
        );
        setPriceData(priceResponse);
      } catch (error) {
        console.error('Ошибка при расчете цены:', error);
        setPriceData(null);
      } finally {
        setPriceLoading(false);
      }
    };

    calculatePrice();
  }, [selectedService, selectedRentalTime, withChemistry]);

  // Обработчик выбора услуги
  const handleServiceSelect = (serviceType) => {
    try {
      setSelectedService(serviceType);
      // Если выбрана услуга без химии, сбрасываем флаг
      if (!serviceType.hasChemistry) {
        setWithChemistry(false);
      }
    } catch (error) {
      console.error('Ошибка в handleServiceSelect:', error);
    }
  };
  
  // Обработчик переключения опции химии
  const handleChemistryToggle = () => {
    try {
      setWithChemistry(!withChemistry);
    } catch (error) {
      console.error('Ошибка в handleChemistryToggle:', error);
    }
  };
  
  // Обработчик выбора времени аренды
  const handleRentalTimeSelect = (time) => {
    try {
      setSelectedRentalTime(time);
    } catch (error) {
      console.error('Ошибка в handleRentalTimeSelect:', error);
    }
  };

  // Обработчик изменения номера машины
  const handleCarNumberChange = (value) => {
    try {
      setCarNumber(value || '');
    } catch (error) {
      console.error('Ошибка в handleCarNumberChange:', error);
      setCarNumber('');
    }
  };

  // Обработчик изменения чекбокса "запомнить номер"
  const handleRememberCarNumberChange = (checked) => {
    try {
      setRememberCarNumber(checked);
    } catch (error) {
      console.error('Ошибка в handleRememberCarNumberChange:', error);
    }
  };

  // Безопасная проверка номера машины (гибкая валидация)
  const isValidCarNumber = (number) => {
    try {
      if (!number || typeof number !== 'string') {
        return false;
      }
      
      // Проверяем минимальную длину
      if (number.length < 6) {
        return false;
      }
      
      // Проверяем, что номер содержит только буквы и цифры
      const carNumberRegex = /^[А-ЯA-Z0-9]+$/;
      if (!carNumberRegex.test(number)) {
        return false;
      }
      
      // Проверяем, что есть хотя бы одна буква и одна цифра
      const hasLetter = /[А-ЯA-Z]/.test(number);
      const hasDigit = /[0-9]/.test(number);
      
      return hasLetter && hasDigit;
    } catch (error) {
      console.error('Ошибка в isValidCarNumber:', error);
      return false;
    }
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
    try {
      if (selectedService && selectedRentalTime && carNumber) {
        // Если пользователь хочет запомнить номер, сохраняем его
        if (rememberCarNumber) {
          await saveCarNumber();
        }

        // Сохраняем данные для платежа
        const serviceData = {
          serviceType: selectedService.id,
          withChemistry: selectedService.hasChemistry ? withChemistry : false,
          rentalTimeMinutes: selectedRentalTime,
          carNumber: carNumber,
          priceData: priceData // Передаем рассчитанную цену
        };
        
        setSelectedServiceData(serviceData);
        setShowPaymentModal(true);
      }
    } catch (error) {
      console.error('Ошибка в handleConfirm:', error);
    }
  };

  // Обработчик успешного платежа
  const handlePaymentSuccess = (payment) => {
    console.log('Платеж успешно завершен:', payment);
    setShowPaymentModal(false);
    onSelect(selectedServiceData);
  };

  // Обработчик ошибки платежа
  const handlePaymentError = (error) => {
    console.error('Ошибка платежа:', error);
    // НЕ закрываем модалку при ошибке - пусть PaymentModal сам обрабатывает ошибки
    // setShowPaymentModal(false);
  };

  // Проверяем, можно ли подтвердить выбор
  const canConfirm = selectedService && 
    selectedRentalTime && 
    carNumber && 
    carNumber.length >= 6 && 
    isValidCarNumber(carNumber);

  // Определяем, нужно ли показывать чекбокс "запомнить номер"
  const showRememberCheckbox = user && 
    carNumber && 
    carNumber !== user.car_number && 
    isValidCarNumber(carNumber);

  // Получаем отображаемую цену
  const getDisplayPrice = () => {
    if (priceLoading) {
      return 'Загрузка...';
    }
    if (priceData) {
      return formatAmount(priceData.total_price_kopecks);
    }
    return '—';
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
      
      {selectedService && (
        <div className={styles.optionsContainer}>
          {/* Ввод номера машины */}
          <CarNumberInput
            value={carNumber || ''}
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
                {(rentalTimes || []).map((time) => (
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
          disabled={!canConfirm || loading || savingCarNumber || priceLoading}
          className={styles.confirmButton}
        >
          {savingCarNumber ? 'Сохранение...' : `Подтвердить выбор (${getDisplayPrice()})`}
        </Button>
      </div>

      {/* Модальное окно платежа */}
      <PaymentModal
        isOpen={showPaymentModal}
        onClose={() => setShowPaymentModal(false)}
        paymentType="queue"
        serviceType={selectedService?.id}
        userID={user?.id}
        rentalTimeMinutes={selectedRentalTime}
        withChemistry={selectedService?.hasChemistry ? withChemistry : false}
        priceData={priceData} // Передаем рассчитанную цену
        onPaymentSuccess={handlePaymentSuccess}
        onPaymentError={handlePaymentError}
      />
    </div>
  );
};

export default ServiceSelector;
