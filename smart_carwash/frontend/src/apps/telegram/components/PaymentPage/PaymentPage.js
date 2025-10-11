import React, { useState, useEffect } from 'react';
import styles from './PaymentPage.module.css';
import { Card, Button } from '../../../../shared/components/UI';
import ApiService from '../../../../shared/services/ApiService';

/**
 * Компонент PaymentPage - страница оплаты услуги
 * @param {Object} props - Свойства компонента
 * @param {Object} props.session - Данные сессии
 * @param {Object} props.payment - Данные платежа
 * @param {Function} props.onPaymentComplete - Функция вызываемая при успешной оплате
 * @param {Function} props.onPaymentFailed - Функция вызываемая при неудачной оплате
 * @param {Function} props.onBack - Функция возврата назад
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {string} props.paymentType - Тип платежа ('main' или 'extension')
 */
const PaymentPage = ({ session, payment, onPaymentComplete, onPaymentFailed, onBack, theme = 'light', paymentType = 'main' }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
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

  // Получение заголовка платежа
  const getPaymentTitle = () => {
    if (paymentType === 'extension') {
      return 'Продление сессии';
    }
    return 'Оплата услуги';
  };

  // Обработка перехода к оплате
  const handlePayment = () => {
    if (payment && payment.payment_url) {
      // Открываем страницу оплаты в новом окне
      window.open(payment.payment_url, '_blank');
      
      // Начинаем проверку статуса платежа
      startPaymentStatusCheck();
    }
  };

  // Проверка статуса платежа
  const startPaymentStatusCheck = () => {
    setLoading(true);
    setError(null);
    
    // Проверяем статус каждые 2 секунды
    const checkInterval = setInterval(async () => {
      try {
        const updatedSession = await ApiService.getUserSession(session.user_id);
        
        if (paymentType === 'extension') {
          // Для продления проверяем, что requested_extension_time_minutes стал 0
          if (updatedSession.session.requested_extension_time_minutes === 0) {
            // Продление успешно применено
            clearInterval(checkInterval);
            setLoading(false);
            onPaymentComplete(updatedSession.session);
          } else if (updatedSession.session.status === 'payment_failed') {
            // Платеж неудачен
            clearInterval(checkInterval);
            setLoading(false);
            onPaymentFailed(updatedSession.session);
          }
        } else {
          // Для основного платежа проверяем статус сессии
          if (updatedSession.session.status === 'in_queue' || updatedSession.session.status === 'assigned') {
            // Платеж успешен
            clearInterval(checkInterval);
            setLoading(false);
            onPaymentComplete(updatedSession.session);
          } else if (updatedSession.session.status === 'payment_failed') {
            // Платеж неудачен
            clearInterval(checkInterval);
            setLoading(false);
            onPaymentFailed(updatedSession.session);
          }
        }
      } catch (err) {
        setError('Ошибка проверки статуса платежа');
        clearInterval(checkInterval);
        setLoading(false);
      }
    }, 2000);

    // Останавливаем проверку через 10 минут
    setTimeout(() => {
      clearInterval(checkInterval);
      setLoading(false);
    }, 600000);
  };

  // Повторная попытка оплаты
  const handleRetryPayment = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await ApiService.retryPayment(session.id);
      
      if (response.payment && response.payment.payment_url) {
        window.open(response.payment.payment_url, '_blank');
        startPaymentStatusCheck();
      }
    } catch (err) {
      console.error('Ошибка повторной оплаты:', err);
      setError('Не удалось создать новый платеж');
      setLoading(false);
    }
  };

  if (!session || !payment) {
    return (
      <div className={`${styles.paymentPage} ${themeClass}`}>
        <Card>
          <div className={styles.error}>
            <h3>Ошибка</h3>
            <p>Данные платежа не найдены</p>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className={`${styles.paymentPage} ${themeClass}`}>
      <Card>
        <div className={styles.header}>
          <h2>{getPaymentTitle()}</h2>
          <p className={styles.subtitle}>
            {paymentType === 'extension' 
              ? 'Подтвердите оплату для продления сессии' 
              : 'Подтвердите оплату для записи в очередь'
            }
          </p>
        </div>

        <div className={styles.paymentInfo}>
          <div className={styles.serviceInfo}>
            <div className={styles.serviceName}>
              {getServiceName(session.service_type)}
              {session.with_chemistry && session.service_type === 'wash' && (
                <span className={styles.chemistryBadge}>+ химия</span>
              )}
            </div>
            <div className={styles.duration}>
              {paymentType === 'extension' 
                ? `${session.requested_extension_time_minutes || 0} минут продления`
                : `${session.rental_time_minutes} минут`
              }
            </div>
            <div className={styles.carNumber}>
              Номер: {session.car_number}
            </div>
          </div>

          <div className={styles.priceInfo}>
            <div className={styles.price}>
              {formatPrice(payment.amount)}
            </div>
            <div className={styles.currency}>
              {payment.currency}
            </div>
          </div>
        </div>

        {error && (
          <div className={styles.error}>
            <p>{error}</p>
          </div>
        )}

        <div className={styles.actions}>
          {loading ? (
            <div className={styles.loading}>
              <p>Ожидание оплаты...</p>
              <p className={styles.loadingHint}>
                Если вы уже оплатили, статус обновится автоматически
              </p>
            </div>
          ) : (
            <>
              <Button 
                onClick={handlePayment}
                disabled={!payment.payment_url}
                className={styles.payButton}
              >
                Перейти к оплате
              </Button>
              
              {session.status === 'payment_failed' && (
                <Button 
                  onClick={handleRetryPayment}
                  variant="secondary"
                  className={styles.retryButton}
                >
                  Повторить оплату
                </Button>
              )}
              
            </>
          )}
        </div>

        <div className={styles.instructions}>
          <h4>Инструкция:</h4>
          <ol>
            <li>Нажмите "Перейти к оплате"</li>
            <li>Заполните данные карты на странице Tinkoff</li>
            <li>Подтвердите оплату</li>
            <li>Вернитесь в приложение - статус обновится автоматически</li>
          </ol>
        </div>
      </Card>
    </div>
  );
};

export default PaymentPage; 