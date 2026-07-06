import React, { useEffect, useState, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styles from './SessionDetails.module.css';
import { Card, Button, StatusBadge, Timer } from '../../../../shared/components/UI';
import { formatDate } from '../../../../shared/utils/formatters';
import { getServiceTypeDescription, formatRefundInfo, formatSessionRefundInfo, formatAmount, formatAmountWithRefund, getPaymentStatusText, getPaymentStatusColor, formatSessionDetailedCost, getDisplayPaymentStatus } from '../../../../shared/utils/statusHelpers';
import ApiService from '../../../../shared/services/ApiService';
import useTimer from '../../../../shared/hooks/useTimer';

// Компонент для отображения статуса и таймера химии
const ChemistryStatus = ({ session }) => {
  const [chemistryTimeLeft, setChemistryTimeLeft] = useState(null);

  // Таймер обратного отсчета химии (если активна)
  useEffect(() => {
    if (!session || !session.chemistry_started_at || session.chemistry_ended_at) {
      setChemistryTimeLeft(null);
      return;
    }

    const updateChemistryTimer = () => {
      const startTime = new Date(session.chemistry_started_at);
      const now = new Date();
      const timeLimit = (session.chemistry_time_minutes || 0) * 60 * 1000;
      const timePassed = now - startTime;
      const remaining = timeLimit - timePassed;

      if (remaining <= 0) {
        setChemistryTimeLeft(0);
      } else {
        setChemistryTimeLeft(Math.floor(remaining / 1000));
      }
    };

    updateChemistryTimer();
    const interval = setInterval(updateChemistryTimer, 1000);

    return () => clearInterval(interval);
  }, [session]);

  // Если химия выключена - не показываем информацию, просто возвращаем null
  if (session.was_chemistry_on && session.chemistry_ended_at) {
    return null;
  }

  // Если химия активна - показываем отдельный таймер
  if (session.was_chemistry_on && session.chemistry_started_at && !session.chemistry_ended_at) {
    return (
      <div style={{ marginTop: '12px' }}>
        <p style={{ fontSize: '13px', fontWeight: 'bold', margin: '0 0 8px 0', color: '#2e7d32' }}>
          🧪 Химия активна:
        </p>
        {chemistryTimeLeft !== null && chemistryTimeLeft > 0 ? (
          <div style={{ 
            padding: '12px', 
            backgroundColor: '#e8f5e9', 
            borderRadius: '8px',
            border: '2px solid #4caf50',
            textAlign: 'center'
          }}>
            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#2e7d32', marginBottom: '4px' }}>
              {Math.floor(chemistryTimeLeft / 60)}:{(chemistryTimeLeft % 60).toString().padStart(2, '0')}
            </div>
            <div style={{ fontSize: '11px', color: '#666' }}>
              до автовыключения
            </div>
          </div>
        ) : (
          <div style={{ 
            padding: '10px', 
            backgroundColor: '#f5f5f5', 
            borderRadius: '6px',
            textAlign: 'center',
            fontSize: '12px',
            color: '#666'
          }}>
            Химия выключается...
          </div>
        )}
      </div>
    );
  }

  return null;
};

// Компонент кнопки включения химии (только кнопка!)
const ChemistryEnableButton = ({ session, theme, onChemistryEnabled }) => {
  const [isEnabling, setIsEnabling] = useState(false);

  const handleEnableChemistry = async () => {
    if (isEnabling) return;

    try {
      setIsEnabling(true);
      await ApiService.enableChemistry(session.id);
      
      // Немедленно обновляем данные сессии для мгновенного отображения изменений
      try {
        const updatedSessionData = await ApiService.getSessionById(session.id);
        if (updatedSessionData && updatedSessionData.session) {
          // Обновляем данные через callback, если он передан
          if (onChemistryEnabled) {
            onChemistryEnabled(updatedSessionData.session, updatedSessionData.payment);
          }
        }
      } catch (refreshError) {
        console.error('Ошибка при обновлении данных сессии:', refreshError);
        // Не показываем ошибку пользователю, поллинг все равно обновит данные
      }
    } catch (error) {
      console.error('Ошибка включения химии:', error);
      alert('Ошибка включения химии: ' + (error.response?.data?.error || error.message));
    } finally {
      setIsEnabling(false);
    }
  };

  return (
    <div style={{ marginTop: '12px' }}>
      <p style={{ fontSize: '12px', color: '#666', marginBottom: '8px' }}>
        Оплачено: {session.chemistry_time_minutes} мин. химии
      </p>
      <Button 
        theme={theme} 
        onClick={handleEnableChemistry}
        disabled={isEnabling}
        style={{ 
          backgroundColor: '#4CAF50',
          color: 'white',
          width: '100%'
        }}
      >
        {isEnabling ? 'Включение...' : '🧪 Включить химию'}
      </Button>
    </div>
  );
};

/**
 * Компонент SessionDetails - отображает детальную информацию о сессии
 * @param {Object} props - Свойства компонента
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Object} props.user - Информация о пользователе
 */
const SessionDetails = ({ theme = 'light', user }) => {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const pathBase = '/telegram';
  
  const [session, setSession] = useState(null);
  const [payment, setPayment] = useState(null);
  const [box, setBox] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [boxChanged, setBoxChanged] = useState(false);
  const [showExtendModal, setShowExtendModal] = useState(false);
  const [showBuyChemistryModal, setShowBuyChemistryModal] = useState(false);
  const [availableRentalTimes, setAvailableRentalTimes] = useState([]);
  const [selectedExtensionTime, setSelectedExtensionTime] = useState(null);
  const [loadingRentalTimes, setLoadingRentalTimes] = useState(false);
  const [availableChemistryTimes, setAvailableChemistryTimes] = useState([]);
  const [selectedChemistryTime, setSelectedChemistryTime] = useState(null);
  const [loadingChemistryTimes, setLoadingChemistryTimes] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  const [sessionPayments, setSessionPayments] = useState(null);
  const [loadingPayments, setLoadingPayments] = useState(false);
  
  // Ref для управления интервалом поллинга
  const pollingInterval = useRef(null);
  
  // Используем хук для таймера
  const { timeLeft } = useTimer(session);
  
  // Проверяем, можно ли отменить сессию
  const canCancelSession = session && ['created', 'in_queue', 'assigned'].includes(session.status);
  
  // Проверяем, можно ли продлить сессию (всегда когда сессия активна и время не истекло)
  const canExtendSession = session && 
    session.status === 'active' && 
    timeLeft !== null && 
    timeLeft > 0; // Время еще не истекло
  
  // Проверяем, можно ли продлить сессию при неуспешной оплате продления
  const canRetryExtension = session && 
    session.status === 'active' && 
    session.requested_extension_time_minutes > 0 && // Запрошено продление
    payment && 
    (payment.status === 'failed' || payment.status === 'pending'); // Но оплата неуспешна
  
  // Проверяем, можно ли докупить химию (химия не была куплена или полностью использована)
  const canBuyChemistry = session && 
    session.status === 'active' && 
    session.service_type === 'wash';
  
  // Получаем информацию о возврате
  const refundInfo = sessionPayments ? formatSessionRefundInfo(sessionPayments) : formatRefundInfo(payment);
  
  // Функция для загрузки платежей сессии
  const loadSessionPayments = async () => {
    if (!sessionId) return;
    
    try {
      setLoadingPayments(true);
      const payments = await ApiService.getSessionPayments(sessionId);
      setSessionPayments(payments);
    } catch (error) {
      console.error('Ошибка при загрузке платежей сессии:', error);
    } finally {
      setLoadingPayments(false);
    }
  };
  
  // Загружаем платежи при изменении сессии
  useEffect(() => {
    if (sessionId) {
      loadSessionPayments();
    }
  }, [sessionId]);

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
    
    // Устанавливаем интервал для поллинга (каждые 2 секунды, но каждую секунду если сессия в очереди)
    const pollInterval = session?.status === 'in_queue' ? 1000 : 2000;
    
    pollingInterval.current = setInterval(async () => {
      try {
        const sessionData = await ApiService.getSessionById(sessionId);
        
        if (sessionData && sessionData.session) {
          setSession(sessionData.session);
          
          // Обновляем информацию о боксе при поллинге - упрощенная логика
          const newBoxNumber = sessionData.session.box_number;
          const currentBoxNumber = box?.number;
          
          if (newBoxNumber) {
            setBox({ number: newBoxNumber });
            // Проверяем, изменился ли номер бокса
            if (currentBoxNumber && currentBoxNumber !== newBoxNumber) {
              setBoxChanged(true);
              // Сбрасываем флаг через 10 секунд
              setTimeout(() => setBoxChanged(false), 10000);
            }
          } else {
            setBox(null);
          }
          
          // Запрашиваем последний актуальный платеж при поллинге
          try {
            const paymentResponse = await ApiService.getUserSessionForPayment(sessionData.session.user_id);
            if (paymentResponse && paymentResponse.payment) {
              setPayment(paymentResponse.payment);
            }
          } catch (paymentError) {
            console.error('Ошибка получения платежа при поллинге:', paymentError);
            // Если не удалось получить платеж, используем платеж из ответа сессии
            if (sessionData.payment) {
              setPayment(sessionData.payment);
            }
          }
          
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
    }, pollInterval);
  };
  
  // Функция для загрузки доступного времени мойки
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
      alert('Ошибка при загрузке доступного времени мойки: ' + err.message);
      setError('Не удалось загрузить доступное время мойки');
    } finally {
      setLoadingRentalTimes(false);
    }
  };

  // Функция для загрузки доступного времени химии
  const fetchAvailableChemistryTimes = async (serviceType) => {
    try {
      setLoadingChemistryTimes(true);
      const response = await ApiService.getAvailableChemistryTimes(serviceType);
      if (response && response.available_chemistry_times) {
        setAvailableChemistryTimes(response.available_chemistry_times);
      }
    } catch (err) {
      console.error('Ошибка при загрузке доступного времени химии:', err);
      setError('Не удалось загрузить доступное время химии');
    } finally {
      setLoadingChemistryTimes(false);
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
      const response = await ApiService.extendSessionWithPayment(sessionId, selectedExtensionTime, selectedChemistryTime);
      
      if (response && response.payment) {
        setPayment(response.payment);
        // Перенаправляем на страницу оплаты
        navigate(`${pathBase}/payment`, { 
          state: { 
            session: response.session,
            payment: response.payment,
            paymentType: 'extension',
            sessionId: sessionId
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

  // Функция для докупки химии
  const handleBuyChemistry = async () => {
    if (!selectedChemistryTime) {
      setError('Выберите время химии');
      return;
    }

    try {
      setActionLoading(true);
      setError(null);
      
      // Вызываем API для докупки химии (ExtensionTimeMinutes = 0, ExtensionChemistryTimeMinutes = selectedChemistryTime)
      const response = await ApiService.extendSessionWithPayment(sessionId, 0, selectedChemistryTime);
      
      if (response && response.payment) {
        setPayment(response.payment);
        // Перенаправляем на страницу оплаты
        navigate(`${pathBase}/payment`, { 
          state: { 
            session: response.session,
            payment: response.payment,
            paymentType: 'extension',
            sessionId: sessionId
          } 
        });
      }
    } catch (err) {
      alert('Ошибка при создании платежа докупки химии: ' + err.message);
      setError('Не удалось создать платеж докупки химии. Пожалуйста, попробуйте еще раз.');
    } finally {
      setActionLoading(false);
    }
  };

  // Функция для открытия модального окна продления
  const openExtendModal = () => {
    if (session && session.service_type) {
      // Сбрасываем предыдущие значения при повторном продлении
      setSelectedExtensionTime(null);
      setSelectedChemistryTime(null);
      
      fetchAvailableRentalTimes(session.service_type);
      // Загружаем время химии только для мойки с химией
      if (session.service_type === 'wash' && session.with_chemistry) {
        fetchAvailableChemistryTimes(session.service_type);
      }
      setShowExtendModal(true);
    }
  };

  // Функция для закрытия модального окна продления
  const closeExtendModal = () => {
    setShowExtendModal(false);
  };

  // Функция для открытия модального окна докупки химии
  const openBuyChemistryModal = () => {
    if (session && session.service_type === 'wash') {
      // Сбрасываем предыдущие значения
      setSelectedChemistryTime(null);
      
      // Загружаем доступное время химии
      fetchAvailableChemistryTimes(session.service_type);
      setShowBuyChemistryModal(true);
    }
  };

  // Функция для закрытия модального окна докупки химии
  const closeBuyChemistryModal = () => {
    setShowBuyChemistryModal(false);
  };

  // Функция для выбора времени продления
  const handleExtensionTimeSelect = (time) => {
    setSelectedExtensionTime(time);
  };

  // Функция для выбора времени химии
  const handleChemistryTimeSelect = (time) => {
    setSelectedChemistryTime(time);
  };

  // Функция для завершения сессии
  const handleCompleteSession = async () => {
    if (!window.confirm('Вы точно уверены, что хотите завершить мойку? Возврата средств не будет.')) return;
    try {
      setActionLoading(true);
      setError(null);
      
      // Вызываем API для завершения сессии
      const response = await ApiService.completeSession(sessionId);
      
      if (response && response.session) {
        // Немедленно обновляем данные сессии для мгновенного отображения изменений
        try {
          const updatedSessionData = await ApiService.getSessionById(sessionId);
          if (updatedSessionData && updatedSessionData.session) {
            setSession(updatedSessionData.session);
            
            // Обновляем информацию о платеже, если она есть
            if (updatedSessionData.payment) {
              setPayment(updatedSessionData.payment);
            }
            
            // Если у сессии есть номер бокса, используем его
            if (updatedSessionData.session.box_number) {
              setBox({ number: updatedSessionData.session.box_number });
            }
            // Иначе, если у сессии есть назначенный бокс, получаем информацию о нем
            else if (updatedSessionData.session.box_id) {
              const queueStatus = await ApiService.getQueueStatus();
              const boxInfo = queueStatus.boxes.find(b => b.id === updatedSessionData.session.box_id);
              if (boxInfo) {
                setBox(boxInfo);
              }
            }
          }
        } catch (refreshError) {
          console.error('Ошибка при обновлении данных сессии:', refreshError);
          // Fallback к данным из ответа API
          setSession(response.session);
          if (response.payment) {
            setPayment(response.payment);
          }
          if (response.session.box_number) {
            setBox({ number: response.session.box_number });
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
        
        // Всегда запрашиваем последний актуальный платеж
        try {
          const paymentResponse = await ApiService.getUserSessionForPayment(response.session.user_id);
          if (paymentResponse && paymentResponse.payment) {
            setPayment(paymentResponse.payment);
          }
        } catch (paymentError) {
          console.error('Ошибка получения платежа:', paymentError);
          // Если не удалось получить платеж, используем платеж из ответа сессии
          if (response.payment) {
            setPayment(response.payment);
          }
        }
        
        // Упрощенная логика обновления бокса - всегда используем box_number
        if (response.session.box_number) {
          setBox({ number: response.session.box_number });
        } else {
          setBox(null);
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

  // Перезапускаем поллинг при изменении статуса сессии для более частого обновления при переназначении
  useEffect(() => {
    if (session?.status) {
      startSessionPolling();
      
      // Если сессия перешла в статус in_queue (переназначение), добавляем дополнительное обновление через 3 секунды
      if (session.status === 'in_queue') {
        const timeoutId = setTimeout(() => {
          fetchSessionDetails(); // Принудительное обновление данных
        }, 3000);
        
        return () => clearTimeout(timeoutId);
      }
    }
  }, [session?.status]);
  
  // Функция для запуска сессии
  const handleStartSession = async () => {
    try {
      setActionLoading(true);
      setError(null);
      
      // Вызываем API для обновления статуса сессии
      const response = await ApiService.startSession(sessionId);
      
      if (response && response.session) {
        // Немедленно обновляем данные сессии для мгновенного отображения изменений
        try {
          const updatedSessionData = await ApiService.getSessionById(sessionId);
          if (updatedSessionData && updatedSessionData.session) {
            setSession(updatedSessionData.session);
            
            // Обновляем информацию о платеже, если она есть
            if (updatedSessionData.payment) {
              setPayment(updatedSessionData.payment);
            }
            
            // Если у сессии есть номер бокса, используем его
            if (updatedSessionData.session.box_number) {
              setBox({ number: updatedSessionData.session.box_number });
            }
            // Иначе, если у сессии есть назначенный бокс, получаем информацию о нем
            else if (updatedSessionData.session.box_id) {
              const queueStatus = await ApiService.getQueueStatus();
              const boxInfo = queueStatus.boxes.find(b => b.id === updatedSessionData.session.box_id);
              if (boxInfo) {
                setBox(boxInfo);
              }
            }
          }
        } catch (refreshError) {
          console.error('Ошибка при обновлении данных сессии:', refreshError);
          // Fallback к данным из ответа API
          setSession(response.session);
          if (response.session.box_number) {
            setBox({ number: response.session.box_number });
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
    navigate(`${pathBase}`);
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
        <Card theme={theme}>
          <p>Загрузка информации о сессии...</p>
        </Card>
      </div>
    );
  }
  
  if (error) {
    return (
      <div className={styles.container}>
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
        <Card theme={theme}>
          <p>Сессия не найдена</p>
          <Button theme={theme} onClick={handleBack}>Вернуться на главную</Button>
        </Card>
      </div>
    );
  }
  
  return (
    <div className={styles.container}>
      <Card theme={theme}>
        <h2 className={`${styles.title} ${themeClass}`}>Информация о сессии</h2>
        <StatusBadge status={session.status} theme={theme} />
        
        {/* Номер бокса с цветным фоном */}
        {(session.box_id || session.box_number) && (
          <div style={{
            marginTop: '12px',
            padding: '12px',
            backgroundColor: '#E3F2FD',
            borderRadius: '8px',
            border: '2px solid #2196F3',
            textAlign: 'center',
            backgroundColor: boxChanged ? '#fff3cd' : '#E3F2FD',
            border: boxChanged ? '2px solid #ffc107' : '2px solid #2196F3',
            transition: 'all 0.3s ease'
          }}>
            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#1976D2' }}>
              Бокс #{session.box_number || 'Неизвестный бокс'}
              {boxChanged && <span style={{ color: '#856404', fontSize: '12px', marginLeft: '8px' }}>🔄 Обновлено!</span>}
            </div>
          </div>
        )}
        
        {/* Таймеры и кнопки под статусом */}
        {session.status === 'active' && timeLeft !== null && (
          <>
            <h2 className={`${styles.title} ${themeClass}`} style={{ marginTop: '20px' }}>Оставшееся время</h2>
            <Timer seconds={timeLeft} theme={theme} />
            
            {/* Таймер химии под таймером времени сессии */}
            {session.with_chemistry && session.was_chemistry_on && (
              <ChemistryStatus session={session} />
            )}
          </>
        )}
        
        {session.status === 'assigned' && timeLeft !== null && (
          <>
            <h2 className={`${styles.title} ${themeClass}`} style={{ marginTop: '20px' }}>Время до старта мойки</h2>
            <Timer seconds={timeLeft} theme={theme} />
          </>
        )}
        
        {/* Кнопки под таймерами */}
        <div style={{ marginTop: '20px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {/* Кнопка включения химии для активной сессии - поднята над продлить время */}
          {session.status === 'active' && 
           session.with_chemistry && 
           session.chemistry_time_minutes > 0 && 
           !session.was_chemistry_on && (
            <ChemistryEnableButton 
              session={session} 
              theme={theme} 
              onChemistryEnabled={(updatedSession, updatedPayment) => {
                // Немедленно обновляем данные сессии
                if (updatedSession) {
                  setSession(updatedSession);
                }
                if (updatedPayment) {
                  setPayment(updatedPayment);
                }
              }}
            />
          )}
          
          {/* Кнопки для активной сессии */}
          {session.status === 'active' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
              {canExtendSession && (
                <Button 
                  theme={theme} 
                  onClick={openExtendModal}
                  disabled={actionLoading}
                  loading={actionLoading}
                  style={{ width: '100%' }}
                >
                  Продлить время
                </Button>
              )}
              {canRetryExtension && (
                <Button 
                  theme={theme} 
                  onClick={openExtendModal}
                  disabled={actionLoading}
                  loading={actionLoading}
                  style={{ width: '100%', backgroundColor: '#FF9800' }}
                >
                  🔄 Повторить продление
                </Button>
              )}
              {canBuyChemistry && (
                <Button 
                  theme={theme} 
                  onClick={openBuyChemistryModal}
                  disabled={actionLoading}
                  loading={actionLoading}
                  style={{ width: '100%', backgroundColor: '#9C27B0' }}
                >
                  🧪 Докупить химию
                </Button>
              )}
              <Button 
                theme={theme} 
                variant="danger"
                onClick={handleCompleteSession}
                disabled={actionLoading}
                loading={actionLoading}
                style={{ width: '100%' }}
              >
                Завершить мойку
              </Button>
            </div>
          )}
          
          {/* Кнопка "Включить бокс" отображается только если сессия в статусе assigned */}
          {session.status === 'assigned' && session.box_id && (
            <Button 
              theme={theme} 
              onClick={handleStartSession}
              disabled={actionLoading}
              loading={actionLoading}
              style={{ width: '100%' }}
            >
              Включить бокс
            </Button>
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
        </div>
        
        {/* Заголовок "Детали сессии" */}
        <h3 className={`${styles.title} ${themeClass}`} style={{ marginTop: '20px', fontSize: '16px' }}>
          Детали сессии
        </h3>
        
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
          <div className={`${styles.infoLabel} ${themeClass}`}>Время мойки:</div>
          <div className={`${styles.infoValue} ${themeClass}`}>
            {session.rental_time_minutes || 5} минут
            {session.extension_time_minutes > 0 && ` (продлено на ${session.extension_time_minutes} минут)`}
            {session.requested_extension_time_minutes > 0 && ` (запрошено продление на ${session.requested_extension_time_minutes} минут)`}
          </div>
        </div>
        
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
                  color: getDisplayPaymentStatus(session) === 'Оплачен' ? '#4CAF50' : 
                         getDisplayPaymentStatus(session) === 'Ожидает оплаты' ? '#FF9800' : 
                         getDisplayPaymentStatus(session) === 'Возвращен' ? '#2196F3' : '#F44336',
                  fontWeight: 'bold'
                }}>
                  {getDisplayPaymentStatus(session)}
                </span>
              </div>
            </div>
            
            {loadingPayments ? (
              <div className={`${styles.infoRow} ${themeClass}`}>
                <div className={`${styles.infoLabel} ${themeClass}`}>Сумма:</div>
                <div className={`${styles.infoValue} ${themeClass}`}>Загрузка...</div>
              </div>
            ) : sessionPayments ? (
              <>
                <div className={`${styles.infoRow} ${themeClass}`}>
                  <div className={`${styles.infoLabel} ${themeClass}`}>Общая стоимость:</div>
                  <div className={`${styles.infoValue} ${themeClass}`}>
                    {formatSessionDetailedCost(sessionPayments).totalCost}
                  </div>
                </div>
                {formatSessionDetailedCost(sessionPayments).details.map((detail, index) => (
                  <div key={index}>
                    <div className={`${styles.infoRow} ${themeClass}`}>
                      <div className={`${styles.infoLabel} ${themeClass}`}>{detail.label}:</div>
                      <div className={`${styles.infoValue} ${themeClass}`}>
                        {detail.value}
                      </div>
                    </div>
                    {detail.refunded && (
                      <div className={`${styles.infoRow} ${themeClass}`}>
                        <div className={`${styles.infoLabel} ${themeClass}`}>Возвращено:</div>
                        <div className={`${styles.infoValue} ${themeClass}`} style={{ color: '#2196F3', fontWeight: 'bold' }}>
                          {detail.refunded}
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </>
            ) : (
              <div className={`${styles.infoRow} ${themeClass}`}>
                <div className={`${styles.infoLabel} ${themeClass}`}>Сумма последнего платежа:</div>
                <div className={`${styles.infoValue} ${themeClass}`}>
                  {formatAmountWithRefund(payment)}
                </div>
              </div>
            )}
            
            {/* Кнопка оплатить для неуспешных платежей (только для основных платежей, не для продлений) */}
            {(payment.status === 'failed' || payment.status === 'pending') && 
             session.status !== 'canceled' && session.status !== 'complete' && session.status !== 'expired' &&
             payment.payment_type === 'main' && (
              <div style={{ marginTop: '20px', textAlign: 'center' }}>
                <Button 
                  theme={theme} 
                  onClick={async () => {
                    try {
                      // Запрашиваем последний платеж по сессии
                      const response = await ApiService.getUserSessionForPayment(session.user_id);
                      
                      navigate(`${pathBase}/payment`, {
                        state: {
                          session: response.session,
                          payment: response.payment,
                          sessionId: session.id
                        }
                      });
                    } catch (error) {
                      console.error('Ошибка получения платежа:', error);
                    }
                  }}
                  style={{ 
                    backgroundColor: '#FF9800',
                    color: 'white',
                    width: '100%'
                  }}
                >
                  💳 Оплатить
                </Button>
              </div>
            )}
          </>
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
                  
                  {/* Выбор времени химии для мойки с химией */}
                  {session && session.service_type === 'wash' && session.with_chemistry && (
                    <>
                      <p className={`${styles.modalText} ${themeClass}`} style={{ marginTop: '20px' }}>
                        Дополнительное время химии (опционально):
                      </p>
                      {loadingChemistryTimes ? (
                        <p className={`${styles.loadingText} ${themeClass}`}>Загрузка времени химии...</p>
                      ) : (
                        <div className={styles.rentalTimeGrid}>
                          {availableChemistryTimes.map((time) => (
                            <div 
                              key={time} 
                              className={`${styles.rentalTimeItem} ${selectedChemistryTime === time ? styles.selectedTime : ''}`}
                              onClick={() => handleChemistryTimeSelect(time)}
                            >
                              <span className={`${styles.rentalTimeValue} ${themeClass}`}>{time}</span>
                              <span className={`${styles.rentalTimeUnit} ${themeClass}`}>мин</span>
                            </div>
                          ))}
                        </div>
                      )}
                    </>
                  )}
                  
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
        
        {/* Модальное окно для докупки химии */}
        {showBuyChemistryModal && (
          <div className={styles.modalOverlay}>
            <Card theme={theme} className={styles.modal}>
              <h3 className={`${styles.modalTitle} ${themeClass}`}>Докупка химии</h3>
              
              {loadingChemistryTimes ? (
                <p className={`${styles.loadingText} ${themeClass}`}>Загрузка доступного времени химии...</p>
              ) : (
                <>
                  <p className={`${styles.modalText} ${themeClass}`}>Выберите время химии:</p>
                  <div className={styles.rentalTimeGrid}>
                    {availableChemistryTimes.map((time) => (
                      <div 
                        key={time} 
                        className={`${styles.rentalTimeItem} ${selectedChemistryTime === time ? styles.selectedTime : ''}`}
                        onClick={() => handleChemistryTimeSelect(time)}
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
                      onClick={closeBuyChemistryModal}
                      disabled={actionLoading}
                    >
                      Отмена
                    </Button>
                    <Button 
                      theme={theme} 
                      onClick={handleBuyChemistry}
                      disabled={actionLoading || !selectedChemistryTime}
                      loading={actionLoading}
                    >
                      Докупить
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
