import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
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
const PaymentPage = ({ session, payment: initialPayment, onPaymentComplete, onPaymentFailed, onBack, theme = 'light', paymentType = 'main' }) => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [paymentFailed, setPaymentFailed] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [payment, setPayment] = useState(initialPayment);
  
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
  const handlePayment = async () => {
    // Если была ошибка оплаты, создаем новый платеж
    if (paymentFailed) {
      await handleRetryPayment();
      return;
    }
    
    // Проверяем, нужно ли создать новый платеж (если это повторная попытка)
    if (retryCount > 0) {
      await handleRetryPayment();
      return;
    }
    
    if (payment && payment.payment_url) {
      // Открываем страницу оплаты в новом окне
      window.open(payment.payment_url, '_blank');
      
      // Начинаем проверку статуса платежа
      startPaymentStatusCheck();
    }
  };

  // Функция для повторной попытки оплаты
  const handleRetryPayment = async () => {
    // Сбрасываем все состояния ошибок
    setPaymentFailed(false);
    setError(null);
    setRetryCount(prev => prev + 1);
    
    try {
      setLoading(true);
      
      // Создаем новый платеж для повторной оплаты
      let newPayment;
      if (paymentType === 'extension') {
        // Для продления создаем новый платеж продления
        const response = await ApiService.extendSessionWithPayment(session.id, session.requested_extension_time_minutes, session.extension_chemistry_time_minutes || 0);
        newPayment = response.payment;
      } else {
        // Для основного платежа создаем новый платеж с той же суммой
        const response = await ApiService.createNewPayment(session.id, payment.amount, payment.currency);
        newPayment = response.payment;
      }
      
      if (newPayment && newPayment.payment_url) {
        // Обновляем payment в состоянии
        setPayment(newPayment);
        
        // Открываем новую ссылку на оплату
        window.open(newPayment.payment_url, '_blank');
        
        // Начинаем проверку статуса нового платежа с его ID
        startPaymentStatusCheck(newPayment.id);
      } else {
        setError('Не удалось создать новый платеж для повторной оплаты');
        setLoading(false);
      }
    } catch (err) {
      setError('Ошибка при создании нового платежа: ' + err.message);
      setLoading(false);
    }
  };

  // Функция для возврата к сессии
  const handleBackToSession = () => {
    if (session?.id) {
      navigate(`/telegram/session/${session.id}`);
    } else {
      onBack();
    }
  };

  // Функция для проверки статуса платежа
  const checkPaymentStatus = async (paymentId = null) => {
    try {
      // Используем переданный ID или ID текущего платежа
      const idToCheck = paymentId || payment.id;
      const response = await ApiService.getPaymentStatus(idToCheck);
      return response.payment;
    } catch (err) {
      console.error('Ошибка проверки статуса платежа:', err);
      return null;
    }
  };

  // Проверка статуса платежа
  const startPaymentStatusCheck = (paymentId = null) => {
    setLoading(true);
    setError(null);
    setPaymentFailed(false);
    
    let checkCount = 0;
    const maxChecks = 30; // 30 проверок по 2 секунды = 1 минута
    
    // Проверяем статус каждые 2 секунды
    const checkInterval = setInterval(async () => {
      try {
        checkCount++;
        
        // Сначала проверяем статус платежа напрямую
        const updatedPayment = await checkPaymentStatus(paymentId || payment.id);
        if (updatedPayment) {
          if (updatedPayment.status === 'succeeded') {
            // Платеж успешен
            clearInterval(checkInterval);
            setLoading(false);
            // Получаем обновленную сессию для передачи в onPaymentComplete
            const updatedSession = await ApiService.getUserSessionForPayment(session.user_id);
            onPaymentComplete(updatedSession.session);
            return;
          } else if (updatedPayment.status === 'failed') {
            // Платеж неудачен
            clearInterval(checkInterval);
            setLoading(false);
            setPaymentFailed(true);
            return;
          }
        }
        
        // Дополнительная проверка через статус сессии
        const updatedSession = await ApiService.getUserSessionForPayment(session.user_id);
        
        if (paymentType === 'extension') {
          // Для продления проверяем, что requested_extension_time_minutes стал 0
          // И что платеж действительно успешен
          if (updatedSession.session.requested_extension_time_minutes === 0 && updatedSession.session.requested_extension_chemistry_time_minutes === 0) {
            // Продление успешно применено
            clearInterval(checkInterval);
            setLoading(false);
            onPaymentComplete(updatedSession.session);
          } else if (checkCount >= maxChecks) {
            // Если прошло много времени без успеха, считаем оплату неудачной
            clearInterval(checkInterval);
            setLoading(false);
            setPaymentFailed(true);
          }
        } else {
          // Для основного платежа проверяем статус сессии
          if (updatedSession.session.status === 'in_queue' || updatedSession.session.status === 'assigned') {
            // Платеж успешен
            clearInterval(checkInterval);
            setLoading(false);
            onPaymentComplete(updatedSession.session);
          } else if (checkCount >= maxChecks) {
            // Если прошло много времени без успеха, считаем оплату неудачной
            clearInterval(checkInterval);
            setLoading(false);
            setPaymentFailed(true);
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
      if (loading) {
        setLoading(false);
        setPaymentFailed(true);
      }
    }, 600000);
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
          ) : (paymentFailed || payment.status === 'failed') ? (
            <div className={styles.paymentFailed}>
              <div className={styles.errorMessage}>
                <h3>❌ Ошибка оплаты</h3>
                <p>Платеж не прошел. Попробуйте еще раз или вернитесь к сессии.</p>
                {retryCount > 0 && (
                  <p className={styles.retryInfo}>
                    Попытка {retryCount + 1}
                  </p>
                )}
              </div>
              <div className={styles.failedActions}>
                <Button 
                  onClick={handleRetryPayment}
                  className={styles.retryButton}
                >
                  🔄 Попробовать снова
                </Button>
                <Button 
                  onClick={handleBackToSession}
                  variant="secondary"
                  className={styles.backButton}
                >
                  ← Вернуться к сессии
                </Button>
              </div>
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
              
            </>
          )}
        </div>

        {/* Показываем инструкцию только если нет ошибки оплаты */}
        {!paymentFailed && payment.status !== 'failed' && (
          <div className={styles.instructions}>
            <h4>Инструкция:</h4>
            <ol>
              <li>Нажмите "Перейти к оплате"</li>
              <li>Заполните данные карты на странице Tinkoff</li>
              <li>Подтвердите оплату</li>
              <li>Вернитесь в приложение - статус обновится автоматически</li>
            </ol>
          </div>
        )}
      </Card>
    </div>
  );
};

export default PaymentPage; 