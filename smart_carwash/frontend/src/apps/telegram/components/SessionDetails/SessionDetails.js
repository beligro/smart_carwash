import React, { useEffect, useState, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styles from './SessionDetails.module.css';
import { Card, Button, StatusBadge, Timer } from '../../../../shared/components/UI';
import { formatDate } from '../../../../shared/utils/formatters';
import { getServiceTypeDescription, formatRefundInfo, formatAmount, formatAmountWithRefund, getPaymentStatusText, getPaymentStatusColor } from '../../../../shared/utils/statusHelpers';
import ApiService from '../../../../shared/services/ApiService';
import useTimer from '../../../../shared/hooks/useTimer';

/**
 * Компонент SessionDetails - отображает детальную информацию о сессии
 * @param {Object} props - Свойства компонента
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Object} props.user - Информация о пользователе
 */
const SessionDetails = ({ theme = 'light', user }) => {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  
  const [session, setSession] = useState(null);
  const [payment, setPayment] = useState(null);
  const [box, setBox] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [showExtendModal, setShowExtendModal] = useState(false);
  const [availableRentalTimes, setAvailableRentalTimes] = useState([]);
  const [selectedExtensionTime, setSelectedExtensionTime] = useState(null);
  const [loadingRentalTimes, setLoadingRentalTimes] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  
  // Ref для управления интервалом поллинга
  const pollingInterval = useRef(null);
  
  // Используем хук для таймера
  const { timeLeft } = useTimer(session);
  
  // Проверяем, можно ли отменить сессию
  const canCancelSession = session && ['created', 'in_queue', 'assigned'].includes(session.status);
  
  // Получаем информацию о возврате
  const refundInfo = formatRefundInfo(payment);

  // Функция для очистки интервала поллинга
  const clearPollingInterval = () => {
    if (pollingInterval.current) {
      clearInterval(pollingInterval.current);
      pollingInterval.current = null;
    }
  };

  // Функция для запуска поллинга сессии
  const startSessionPolling = () => {
    // Очищаем старый интервал, если он существует
    clearPollingInterval();
    
    // Устанавливаем интервал для поллинга (каждые 5 секунд)
    pollingInterval.current = setInterval(async () => {
      try {
        const sessionData = await ApiService.getSessionById(sessionId);
        
        if (sessionData && sessionData.session) {
          setSession(sessionData.session);
          setPayment(sessionData.payment);
          
          // Если сессия завершена, отменена или истекла, останавливаем поллинг
          if (
            sessionData.session.status === 'complete' || 
            sessionData.session.status === 'canceled' ||
            sessionData.session.status === 'expired'
          ) {
            clearPollingInterval();
          }
        }
      } catch (err) {
        console.error('Ошибка при поллинге сессии:', err);
        // Не показываем ошибку пользователю, просто логируем
      }
    }, 5000);
  };
  
  // Функция для загрузки доступного времени аренды
  const fetchAvailableRentalTimes = async (serviceType) => {
    try {
      setLoadingRentalTimes(true);
      const response = await ApiService.getAvailableRentalTimes(serviceType);
      if (response && response.available_times) {
        setAvailableRentalTimes(response.available_times);
        // Устанавливаем первое значение как выбранное по умолчанию
        if (response.available_times.length > 0) {
          setSelectedExtensionTime(response.available_times[0]);
        }
      }
    } catch (err) {
      alert('Ошибка при загрузке доступного времени аренды: ' + err.message);
      setError('Не удалось загрузить доступное время аренды');
    } finally {
      setLoadingRentalTimes(false);
    }
  };

  // Функция для продления сессии с оплатой
  const handleExtendSession = async () => {
    if (!selectedExtensionTime) {
      setError('Выберите время продления');
      return;
    }

    try {
      setActionLoading(true);
      setError(null);
      
      // Вызываем API для продления сессии с оплатой
      const response = await ApiService.extendSessionWithPayment(sessionId, selectedExtensionTime);
      
      if (response && response.payment) {
        // Перенаправляем на страницу оплаты
        navigate('/telegram/payment', { 
          state: { 
            session: response.session,
            payment: response.payment,
            paymentType: 'extension'
          } 
        });
      }
    } catch (err) {
      alert('Ошибка при создании платежа продления: ' + err.message);
      setError('Не удалось создать платеж продления. Пожалуйста, попробуйте еще раз.');
    } finally {
      setActionLoading(false);
    }
  };

  // Функция для открытия модального окна продления
  const openExtendModal = () => {
    if (session && session.service_type) {
      fetchAvailableRentalTimes(session.service_type);
      setShowExtendModal(true);
    }
  };

  // Функция для закрытия модального окна продления
  const closeExtendModal = () => {
    setShowExtendModal(false);
  };

  // Функция для выбора времени продления
  const handleExtensionTimeSelect = (time) => {
    setSelectedExtensionTime(time);
  };

  // Функция для завершения сессии
  const handleCompleteSession = async () => {
    try {
      setActionLoading(true);
      setError(null);
      
      // Вызываем API для завершения сессии
      const response = await ApiService.completeSession(sessionId);
      
      if (response && response.session) {
        setSession(response.session);
        
        // Обновляем информацию о платеже, если она есть
        if (response.payment) {
          setPayment(response.payment);
        }
        
        // Если у сессии есть номер бокса, используем его
        if (response.session.box_number) {
          setBox({ number: response.session.box_number });
        }
        // Иначе, если у сессии есть назначенный бокс, получаем информацию о нем
        else if (response.session.box_id) {
          const queueStatus = await ApiService.getQueueStatus();
          const boxInfo = queueStatus.boxes.find(b => b.id === response.session.box_id);
          if (boxInfo) {
            setBox(boxInfo);
          }
        }
      }
    } catch (err) {
      alert('Ошибка при завершении сессии: ' + err.message);
      setError('Не удалось завершить сессию. Пожалуйста, попробуйте еще раз.');
    } finally {
      setActionLoading(false);
    }
  };
  
  // Функция для загрузки данных о сессии
  const fetchSessionDetails = async () => {
    try {
      setLoading(true);
      const response = await ApiService.getSessionById(sessionId);
      
      if (response && response.session) {
        setSession(response.session);
        if (response.payment) {
          setPayment(response.payment);
        }
        
        // Если у сессии есть номер бокса, используем его
        if (response.session.box_number) {
          setBox({ number: response.session.box_number });
        }
        // Иначе, если у сессии есть назначенный бокс, получаем информацию о нем
        else if (response.session.box_id) {
          const queueStatus = await ApiService.getQueueStatus();
          const boxInfo = queueStatus.boxes.find(b => b.id === response.session.box_id);
          if (boxInfo) {
            setBox(boxInfo);
          }
        }
        
        return response.session;
      } else {
        setError('Сессия не найдена');
        return null;
      }
    } catch (err) {
      alert('Ошибка при загрузке данных о сессии: ' + err.message);
      setError('Не удалось загрузить данные о сессии');
      return null;
    } finally {
      setLoading(false);
    }
  };
  
  // Начальная загрузка данных о сессии и запуск поллинга
  useEffect(() => {
    if (sessionId) {
      fetchSessionDetails();
      startSessionPolling(); // Запускаем поллинг при первой загрузке
    }
    
    // Очистка интервала при размонтировании компонента
    return () => {
      clearPollingInterval();
    };
  }, [sessionId]);
  
  // Функция для запуска сессии
  const handleStartSession = async () => {
    try {
      setActionLoading(true);
      setError(null);
      
      // Вызываем API для обновления статуса сессии
      const response = await ApiService.startSession(sessionId);
      
      if (response && response.session) {
        setSession(response.session);
        
        // Если у сессии есть номер бокса, используем его
        if (response.session.box_number) {
          setBox({ number: response.session.box_number });
        }
        // Иначе, если у сессии есть назначенный бокс, получаем информацию о нем
        else if (response.session.box_id) {
          const queueStatus = await ApiService.getQueueStatus();
          const boxInfo = queueStatus.boxes.find(b => b.id === response.session.box_id);
          if (boxInfo) {
            setBox(boxInfo);
          }
        }
      }
    } catch (err) {
      alert('Ошибка при запуске сессии: ' + err.message);
      setError('Не удалось запустить сессию. Пожалуйста, попробуйте еще раз.');
    } finally {
      setActionLoading(false);
    }
  };
  
  // Функция для возврата на главную страницу
  const handleBack = () => {
    // Всегда возвращаемся на главную страницу
    navigate('/telegram');
  };

  // Обработчик отмены сессии
  const handleCancelSession = async () => {
    if (!session || !user) return;
    
    const confirmMessage = refundInfo.hasRefund 
      ? `Вы уверены, что хотите отменить сессию? Деньги в размере ${formatAmountWithRefund(payment)} будут возвращены на карту.`
      : 'Вы уверены, что хотите отменить сессию?';
    
    if (!window.confirm(confirmMessage)) return;
    
    try {
      setIsCanceling(true);
      const response = await ApiService.cancelSession(session.id, user.id);
      
      // Обновляем информацию о сессии
      setSession(response.session);
      setPayment(response.payment);
      
      alert('Сессия успешно отменена' + (refundInfo.hasRefund ? '. Деньги будут возвращены на карту.' : ''));
    } catch (error) {
      alert('Ошибка при отмене сессии: ' + error.message);
    } finally {
      setIsCanceling(false);
    }
  };

  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  if (loading) {
    return (
      <div className={styles.container}>
        <button className={`${styles.backButton} ${themeClass}`} onClick={handleBack}>← Назад</button>
        <Card theme={theme}>
          <p>Загрузка информации о сессии...</p>
        </Card>
      </div>
    );
  }
  
  if (error) {
    return (
      <div className={styles.container}>
        <button className={`${styles.backButton} ${themeClass}`} onClick={handleBack}>← Назад</button>
        <Card theme={theme}>
          <div className={`${styles.errorMessage} ${themeClass}`}>{error}</div>
          <Button theme={theme} onClick={handleBack}>Вернуться на главную</Button>
        </Card>
      </div>
    );
  }
  
  if (!session) {
    return (
      <div className={styles.container}>
        <button className={`${styles.backButton} ${themeClass}`} onClick={handleBack}>← Назад</button>
        <Card theme={theme}>
          <p>Сессия не найдена</p>
          <Button theme={theme} onClick={handleBack}>Вернуться на главную</Button>
        </Card>
      </div>
    );
  }
  
  return (
    <div className={styles.container}>
      <button className={`${styles.backButton} ${themeClass}`} onClick={handleBack}>← Назад</button>
      
      <Card theme={theme}>
        <h2 className={`${styles.title} ${themeClass}`}>Информация о сессии</h2>
        <StatusBadge status={session.status} theme={theme} />
        
        <div className={`${styles.infoRow} ${themeClass}`}>
          <div className={`${styles.infoLabel} ${themeClass}`}>ID сессии:</div>
          <div className={`${styles.infoValue} ${themeClass}`}>{session.id}</div>
        </div>
        
        <div className={`${styles.infoRow} ${themeClass}`}>
          <div className={`${styles.infoLabel} ${themeClass}`}>Создана:</div>
          <div className={`${styles.infoValue} ${themeClass}`}>{formatDate(session.created_at)}</div>
        </div>
        
        <div className={`${styles.infoRow} ${themeClass}`}>
          <div className={`${styles.infoLabel} ${themeClass}`}>Обновлена:</div>
          <div className={`${styles.infoValue} ${themeClass}`}>{formatDate(session.updated_at)}</div>
        </div>
        
        <div className={`${styles.infoRow} ${themeClass}`}>
          <div className={`${styles.infoLabel} ${themeClass}`}>Тип услуги:</div>
          <div className={`${styles.infoValue} ${themeClass}`}>
            {getServiceTypeDescription(session.service_type)}
            {session.with_chemistry && ' (с химией)'}
          </div>
        </div>
        
        <div className={`${styles.infoRow} ${themeClass}`}>
          <div className={`${styles.infoLabel} ${themeClass}`}>Номер машины:</div>
          <div className={`${styles.infoValue} ${themeClass}`}>{session.car_number || 'Не указан'}</div>
        </div>
        
        <div className={`${styles.infoRow} ${themeClass}`}>
          <div className={`${styles.infoLabel} ${themeClass}`}>Время аренды:</div>
          <div className={`${styles.infoValue} ${themeClass}`}>
            {session.rental_time_minutes || 5} минут
            {session.extension_time_minutes > 0 && ` (продлено на ${session.extension_time_minutes} минут)`}
          </div>
        </div>
        
        {(session.box_id || session.box_number) && (
          <div className={`${styles.infoRow} ${themeClass}`}>
            <div className={`${styles.infoLabel} ${themeClass}`}>Назначенный бокс:</div>
            <div className={`${styles.infoValue} ${themeClass}`}>
              {box ? `Бокс #${box.number}` : 
               session.box_number ? `Бокс #${session.box_number}` : 
               'Информация о боксе недоступна'}
            </div>
          </div>
        )}
        
        {/* Информация о платеже */}
        {payment && (
          <>
            <h3 className={`${styles.title} ${themeClass}`} style={{ marginTop: '20px', fontSize: '16px' }}>
              Информация об оплате
            </h3>
            
            <div className={`${styles.infoRow} ${themeClass}`}>
              <div className={`${styles.infoLabel} ${themeClass}`}>Статус платежа:</div>
              <div className={`${styles.infoValue} ${themeClass}`}>
                <span style={{ 
                  color: payment.status === 'succeeded' ? '#4CAF50' : 
                         payment.status === 'pending' ? '#FF9800' : 
                         payment.status === 'refunded' ? '#2196F3' : '#F44336',
                  fontWeight: 'bold'
                }}>
                  {getPaymentStatusText(payment.status)}
                </span>
              </div>
            </div>
            
            <div className={`${styles.infoRow} ${themeClass}`}>
              <div className={`${styles.infoLabel} ${themeClass}`}>Сумма:</div>
              <div className={`${styles.infoValue} ${themeClass}`}>
                {formatAmountWithRefund(payment)}
              </div>
            </div>
            
            {refundInfo.hasRefund && (
              <>
                <div className={`${styles.infoRow} ${themeClass}`}>
                  <div className={`${styles.infoLabel} ${themeClass}`}>Возвращено:</div>
                  <div className={`${styles.infoValue} ${themeClass}`} style={{ color: '#2196F3', fontWeight: 'bold' }}>
                    {formatAmount(refundInfo.refundedAmount)}
                    {refundInfo.refundType === 'partial' && ` (частично)`}
                    {refundInfo.refundType === 'full' && ` (полностью)`}
                  </div>
                </div>
                
                <div className={`${styles.infoRow} ${themeClass}`}>
                  <div className={`${styles.infoLabel} ${themeClass}`}>Итого:</div>
                  <div className={`${styles.infoValue} ${themeClass}`} style={{ color: '#2196F3', fontWeight: 'bold' }}>
                    {formatAmount(refundInfo.finalAmount)}
                  </div>
                </div>
                
                {payment.refunded_at && (
                  <div className={`${styles.infoRow} ${themeClass}`}>
                    <div className={`${styles.infoLabel} ${themeClass}`}>Время возврата:</div>
                    <div className={`${styles.infoValue} ${themeClass}`}>
                      {formatDate(payment.refunded_at)}
                    </div>
                  </div>
                )}
              </>
            )}
            
            {/* Кнопка повторной оплаты для неудачных платежей */}
            {payment.status === 'failed' && (
              <Button 
                theme={theme} 
                onClick={() => {
                  navigate('/telegram/payment', {
                    state: {
                      session: session,
                      payment: payment
                    }
                  });
                }}
                style={{ 
                  marginTop: '12px',
                  backgroundColor: '#F44336',
                  color: 'white'
                }}
              >
                🔄 Повторить оплату
              </Button>
            )}
          </>
        )}
        
        {/* Таймер отображается для активной сессии */}
        {session.status === 'active' && timeLeft !== null && (
          <>
            <h2 className={`${styles.title} ${themeClass}`} style={{ marginTop: '20px' }}>Оставшееся время мойки</h2>
            <Timer seconds={timeLeft} theme={theme} />
          </>
        )}
        
        {/* Таймер отображается для назначенной сессии */}
        {session.status === 'assigned' && timeLeft !== null && (
          <>
            <h2 className={`${styles.title} ${themeClass}`} style={{ marginTop: '20px' }}>Время до истечения резерва</h2>
            <Timer seconds={timeLeft} theme={theme} />
            <p style={{ textAlign: 'center', marginTop: '10px', color: timeLeft <= 60 ? '#C62828' : 'inherit' }}>
              Начните мойку до истечения времени, иначе резерв будет снят
            </p>
          </>
        )}
        
        {/* Кнопка "Старт сессии" отображается только если сессия в статусе assigned */}
        {session.status === 'assigned' && session.box_id && (
          <Button 
            theme={theme} 
            onClick={handleStartSession}
            disabled={actionLoading}
            loading={actionLoading}
          >
            Начать мойку
          </Button>
        )}
        
        {/* Кнопки для активной сессии */}
        {session.status === 'active' && (
          <div className={styles.buttonGroup}>
            <Button 
              theme={theme} 
              onClick={openExtendModal}
              disabled={actionLoading}
              loading={actionLoading}
              style={{ marginTop: '10px', marginRight: '10px' }}
            >
              Продлить мойку
            </Button>
            <Button 
              theme={theme} 
              variant="danger"
              onClick={handleCompleteSession}
              disabled={actionLoading}
              loading={actionLoading}
              style={{ marginTop: '10px' }}
            >
              Завершить мойку
            </Button>
          </div>
        )}

        {/* Кнопка отмены сессии */}
        {canCancelSession && (
          <Button 
            theme={theme} 
            onClick={handleCancelSession}
            disabled={isCanceling}
            loading={isCanceling}
            style={{ 
              marginTop: '12px',
              backgroundColor: '#F44336',
              color: 'white'
            }}
          >
            {isCanceling ? 'Отмена...' : 'Отменить сессию'}
          </Button>
        )}
        
        {/* Модальное окно для продления сессии */}
        {showExtendModal && (
          <div className={styles.modalOverlay}>
            <Card theme={theme} className={styles.modal}>
              <h3 className={`${styles.modalTitle} ${themeClass}`}>Продление сессии</h3>
              
              {loadingRentalTimes ? (
                <p className={`${styles.loadingText} ${themeClass}`}>Загрузка доступного времени...</p>
              ) : (
                <>
                  <p className={`${styles.modalText} ${themeClass}`}>Выберите время продления:</p>
                  <div className={styles.rentalTimeGrid}>
                    {availableRentalTimes.map((time) => (
                      <div 
                        key={time} 
                        className={`${styles.rentalTimeItem} ${selectedExtensionTime === time ? styles.selectedTime : ''}`}
                        onClick={() => handleExtensionTimeSelect(time)}
                      >
                        <span className={`${styles.rentalTimeValue} ${themeClass}`}>{time}</span>
                        <span className={`${styles.rentalTimeUnit} ${themeClass}`}>мин</span>
                      </div>
                    ))}
                  </div>
                  
                  <div className={styles.modalButtons}>
                    <Button 
                      theme={theme} 
                      variant="secondary"
                      onClick={closeExtendModal}
                      disabled={actionLoading}
                    >
                      Отмена
                    </Button>
                    <Button 
                      theme={theme} 
                      onClick={handleExtendSession}
                      disabled={actionLoading || !selectedExtensionTime}
                      loading={actionLoading}
                    >
                      Продлить
                    </Button>
                  </div>
                </>
              )}
              
              {error && <div className={`${styles.errorMessage} ${themeClass}`}>{error}</div>}
            </Card>
          </div>
        )}
        
        {error && <div className={`${styles.errorMessage} ${themeClass}`}>{error}</div>}
      </Card>
    </div>
  );
};

export default SessionDetails;
