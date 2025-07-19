import React, { useState, useEffect } from 'react';
import PaymentService, { formatAmount } from '../../../../shared/services/PaymentService';
import styles from './PaymentModal.module.css';

const PaymentModal = ({ 
  isOpen, 
  onClose, 
  paymentType, 
  serviceType, 
  sessionID, 
  extensionMinutes,
  userID,
  priceData: initialPriceData, // Принимаем предрассчитанную цену
  onPaymentSuccess,
  onPaymentError 
}) => {
  const [loading, setLoading] = useState(false);
  const [payment, setPayment] = useState(null);
  const [error, setError] = useState(null);
  const [priceData, setPriceData] = useState(null);
  const [priceLoading, setPriceLoading] = useState(false);

  // Используем предрассчитанную цену или рассчитываем заново
  useEffect(() => {
    if (!isOpen) {
      setPriceData(null);
      return;
    }

    if (initialPriceData) {
      // Используем предрассчитанную цену
      setPriceData(initialPriceData);
      setPriceLoading(false);
    } else {
      // Рассчитываем цену самостоятельно (для случаев, когда цена не передана)
      const calculatePrice = async () => {
        setPriceLoading(true);
        try {
          let priceResponse;
          
          if (paymentType === 'queue') {
            // Для очереди используем базовые параметры
            priceResponse = await PaymentService.calculatePrice(
              serviceType,
              5, // базовое время аренды
              false // без химии по умолчанию
            );
          } else if (paymentType === 'extension') {
            // Для продления используем время продления
            priceResponse = await PaymentService.calculatePrice(
              serviceType,
              extensionMinutes,
              false // без химии для продления
            );
          }
          
          setPriceData(priceResponse);
        } catch (error) {
          console.error('Ошибка при расчете цены:', error);
          setPriceData(null);
        } finally {
          setPriceLoading(false);
        }
      };

      calculatePrice();
    }
  }, [isOpen, paymentType, serviceType, extensionMinutes, initialPriceData]);

  // Получаем данные для платежа
  const getPaymentData = () => {
    const getDisplayPrice = () => {
      if (priceLoading) {
        return 'Загрузка...';
      }
      if (priceData) {
        return formatAmount(priceData.total_price_kopecks);
      }
      return '—';
    };

    switch (paymentType) {
      case 'queue':
        return {
          title: 'Оплата за очередь',
          description: `Стоимость услуги: ${getDisplayPrice()}`,
          serviceType,
          userID,
        };
      case 'extension':
        return {
          title: 'Продление сессии',
          description: `Продление на ${extensionMinutes} минут: ${getDisplayPrice()}`,
          sessionID,
          extensionMinutes,
        };
      default:
        return null;
    }
  };

  // Создание платежа
  const createPayment = async () => {
    setLoading(true);
    setError(null);

    try {
      let response;
      
      if (paymentType === 'queue') {
        response = await PaymentService.createQueuePayment(userID, serviceType);
      } else if (paymentType === 'extension') {
        response = await PaymentService.createSessionExtensionPayment(sessionID, extensionMinutes);
      }

      setPayment(response.payment);
      
      // Открываем платежную форму в Telegram
      if (window.Telegram?.WebApp) {
        window.Telegram.WebApp.openTelegramLink(response.payment_url);
      } else {
        // Fallback для тестирования
        window.open(response.payment_url, '_blank');
      }

      onPaymentSuccess?.(response.payment);
    } catch (err) {
      setError(err.response?.data?.error || 'Ошибка создания платежа');
      onPaymentError?.(err);
    } finally {
      setLoading(false);
    }
  };

  // Проверка статуса платежа
  const checkPaymentStatus = async (paymentID) => {
    try {
      const response = await PaymentService.getPaymentStatus(paymentID);
      return response.status;
    } catch (err) {
      console.error('Ошибка проверки статуса платежа:', err);
      return null;
    }
  };

  // Обработка успешного платежа
  const handlePaymentSuccess = () => {
    onPaymentSuccess?.(payment);
    onClose();
  };

  // Обработка неудачного платежа
  const handlePaymentError = () => {
    setError('Платеж не был завершен');
    onPaymentError?.();
  };

  // Обработка закрытия платежа
  const handlePaymentClose = () => {
    onClose();
  };

  useEffect(() => {
    if (isOpen && window.Telegram?.WebApp) {
      // Настраиваем обработчики Telegram WebApp
      window.Telegram.WebApp.onEvent('paymentSuccess', handlePaymentSuccess);
      window.Telegram.WebApp.onEvent('paymentError', handlePaymentError);
      window.Telegram.WebApp.onEvent('paymentClose', handlePaymentClose);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const paymentData = getPaymentData();
  if (!paymentData) return null;

  return (
    <div className={styles.overlay}>
      <div className={styles.modal}>
        <div className={styles.header}>
          <h2>{paymentData.title}</h2>
          <button className={styles.closeButton} onClick={onClose}>
            ✕
          </button>
        </div>

        <div className={styles.content}>
          <p className={styles.description}>{paymentData.description}</p>

          {error && (
            <div className={styles.error}>
              {error}
            </div>
          )}

          {payment ? (
            <div className={styles.paymentInfo}>
              <p>Платеж создан. Открывается форма оплаты...</p>
              <div className={styles.paymentStatus}>
                Статус: {payment.status}
              </div>
            </div>
          ) : (
            <div className={styles.actions}>
              <button
                className={styles.payButton}
                onClick={createPayment}
                disabled={loading || priceLoading}
              >
                {loading ? 'Создание платежа...' : 'Оплатить'}
              </button>
              
              <button
                className={styles.cancelButton}
                onClick={onClose}
                disabled={loading}
              >
                Отмена
              </button>
            </div>
          )}
        </div>

        <div className={styles.footer}>
          <p className={styles.disclaimer}>
            Оплата производится через Tinkoff Kassa. 
            После успешной оплаты вы будете автоматически добавлены в очередь.
          </p>
        </div>
      </div>
    </div>
  );
};

export default PaymentModal; 