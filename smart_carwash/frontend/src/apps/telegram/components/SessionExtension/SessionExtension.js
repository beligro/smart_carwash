import React, { useState, useEffect } from 'react';
import styles from './SessionExtension.module.css';
import { Card, Button } from '../../../../shared/components/UI';
import PaymentModal from '../PaymentModal';
import PaymentService, { formatAmount } from '../../../../shared/services/PaymentService';

const SessionExtension = ({ 
  session, 
  onExtension, 
  theme = 'light',
  user 
}) => {
  const [selectedMinutes, setSelectedMinutes] = useState(5);
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [loading, setLoading] = useState(false);
  const [priceData, setPriceData] = useState(null);
  const [priceLoading, setPriceLoading] = useState(false);

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  // Доступные варианты продления
  const extensionOptions = [5, 10, 15, 20, 30];

  // Расчет цены при изменении времени продления
  useEffect(() => {
    const calculatePrice = async () => {
      if (!selectedMinutes) {
        setPriceData(null);
        return;
      }

      setPriceLoading(true);
      try {
        // Для продления используем тип услуги из сессии
        const priceResponse = await PaymentService.calculatePrice(
          session.service_type,
          selectedMinutes,
          session.with_chemistry || false
        );
        setPriceData(priceResponse);
      } catch (error) {
        console.error('Ошибка при расчете цены продления:', error);
        setPriceData(null);
      } finally {
        setPriceLoading(false);
      }
    };

    calculatePrice();
  }, [selectedMinutes, session.service_type, session.with_chemistry]);

  // Обработчик выбора времени продления
  const handleTimeSelect = (minutes) => {
    setSelectedMinutes(minutes);
  };

  // Обработчик подтверждения продления
  const handleConfirmExtension = () => {
    setShowPaymentModal(true);
  };

  // Обработчик успешного платежа
  const handlePaymentSuccess = (payment) => {
    console.log('Платеж за продление успешно завершен:', payment);
    setShowPaymentModal(false);
    onExtension(selectedMinutes);
  };

  // Обработчик ошибки платежа
  const handlePaymentError = (error) => {
    console.error('Ошибка платежа за продление:', error);
    setShowPaymentModal(false);
  };

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
      <h2 className={`${styles.title} ${themeClass}`}>Продление сессии</h2>
      
      <Card theme={theme} className={styles.infoCard}>
        <div className={styles.sessionInfo}>
          <h3 className={`${styles.sessionTitle} ${themeClass}`}>
            Текущая сессия
          </h3>
          <div className={styles.sessionDetails}>
            <p className={`${styles.sessionDetail} ${themeClass}`}>
              <strong>Номер машины:</strong> {session.car_number}
            </p>
            <p className={`${styles.sessionDetail} ${themeClass}`}>
              <strong>Услуга:</strong> {session.service_type}
            </p>
            <p className={`${styles.sessionDetail} ${themeClass}`}>
              <strong>Осталось времени:</strong> {session.remaining_minutes} мин
            </p>
          </div>
        </div>
      </Card>

      <Card theme={theme} className={styles.extensionCard}>
        <h3 className={`${styles.extensionTitle} ${themeClass}`}>
          Выберите время продления
        </h3>
        
        <div className={styles.extensionGrid}>
          {extensionOptions.map((minutes) => (
            <div
              key={minutes}
              className={`${styles.extensionOption} ${
                selectedMinutes === minutes ? styles.selected : ''
              }`}
              onClick={() => handleTimeSelect(minutes)}
            >
              <span className={`${styles.extensionTime} ${themeClass}`}>
                +{minutes} мин
              </span>
              <span className={`${styles.extensionPrice} ${themeClass}`}>
                {priceLoading ? '...' : (priceData && selectedMinutes === minutes ? formatAmount(priceData.total_price_kopecks) : '—')}
              </span>
            </div>
          ))}
        </div>

        <div className={styles.totalInfo}>
          <p className={`${styles.totalText} ${themeClass}`}>
            Стоимость продления: <strong>{getDisplayPrice()}</strong>
          </p>
          <p className={`${styles.totalDescription} ${themeClass}`}>
            После продления у вас будет {session.remaining_minutes + selectedMinutes} минут
          </p>
        </div>
      </Card>

      <div className={styles.buttonContainer}>
        <Button
          theme={theme}
          onClick={handleConfirmExtension}
          disabled={loading || priceLoading}
          className={styles.extensionButton}
        >
          {loading ? 'Обработка...' : `Продлить на ${selectedMinutes} мин`}
        </Button>
      </div>

      {/* Модальное окно платежа */}
      <PaymentModal
        isOpen={showPaymentModal}
        onClose={() => setShowPaymentModal(false)}
        paymentType="extension"
        sessionID={session.id}
        extensionMinutes={selectedMinutes}
        userID={user?.id}
        priceData={priceData} // Передаем рассчитанную цену
        onPaymentSuccess={handlePaymentSuccess}
        onPaymentError={handlePaymentError}
      />
    </div>
  );
};

export default SessionExtension; 