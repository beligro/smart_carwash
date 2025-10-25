import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './WashInfo.module.css';
import { Card, Button, StatusBadge, Timer } from '../../../../shared/components/UI';
import { formatDate } from '../../../../shared/utils/formatters';
import { getSessionStatusDescription, getServiceTypeDescription, formatRefundInfo, formatAmount, formatAmountWithRefund, getPaymentStatusText, getPaymentStatusColor, formatSessionTotalCost, formatSessionDetailedCost } from '../../../../shared/utils/statusHelpers';
import useTimer from '../../../../shared/hooks/useTimer';
import ApiService from '../../../../shared/services/ApiService';
// import { useSettings } from '../../../../shared/contexts/SettingsContext';

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

  // Если химия выключена
  if (session.was_chemistry_on && session.chemistry_ended_at) {
    return (
      <div style={{ marginTop: '8px', padding: '10px', backgroundColor: '#f5f5f5', borderRadius: '6px', fontSize: '12px', color: '#666' }}>
        ✓ Химия была использована ({session.chemistry_time_minutes} мин)
      </div>
    );
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

/**
 * Вспомогательная функция для форматирования текста очереди с временем ожидания
 * @param {Object} queueInfo - Информация об очереди
 * @returns {string} - Отформатированный текст
 */
const formatQueueText = (queueInfo) => {
  if (!queueInfo.has_queue) {
    return 'Нет очереди';
  }
  
  const baseText = `В очереди: ${queueInfo.queue_size}`;
  
  // Если есть время ожидания, добавляем его
  if (queueInfo.wait_time_minutes && queueInfo.wait_time_minutes > 0) {
    return `${baseText} (ожидание ~${queueInfo.wait_time_minutes} мин)`;
  }
  
  return baseText;
};

/**
 * Компонент WashInfo - отображает информацию о мойке
 * @param {Object} props - Свойства компонента
 * @param {Object} props.washInfo - Информация о мойке
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Function} props.onCreateSession - Функция для создания сессии
 * @param {Function} props.onViewHistory - Функция для просмотра истории сессий
 * @param {Function} props.onCancelSession - Функция для отмены сессии
 * @param {Object} props.user - Данные пользователя
 */
const WashInfo = ({ washInfo, theme = 'light', onCreateSession, onViewHistory, onCancelSession, onChemistryEnabled, onCompleteSession, user }) => {
  const navigate = useNavigate();
  const [isCanceling, setIsCanceling] = useState(false);
  const [sessionPayments, setSessionPayments] = useState(null);
  const [loadingPayments, setLoadingPayments] = useState(false);
  const [boxChanged, setBoxChanged] = useState(false);
  
  // Состояния для модальных окон и действий
  const [actionLoading, setActionLoading] = useState(false);
  const [showExtendModal, setShowExtendModal] = useState(false);
  const [showBuyChemistryModal, setShowBuyChemistryModal] = useState(false);
  const [availableRentalTimes, setAvailableRentalTimes] = useState([]);
  const [selectedExtensionTime, setSelectedExtensionTime] = useState(null);
  const [loadingRentalTimes, setLoadingRentalTimes] = useState(false);
  const [availableChemistryTimes, setAvailableChemistryTimes] = useState([]);
  const [selectedChemistryTime, setSelectedChemistryTime] = useState(null);
  const [loadingChemistryTimes, setLoadingChemistryTimes] = useState(false);
  
  // Кэш для настроек (загружаем один раз)
  const [allChemistryTimesFromSettings, setAllChemistryTimesFromSettings] = useState([]);
  const [allRentalTimesFromSettings, setAllRentalTimesFromSettings] = useState([]);
  
  // Получаем данные из washInfo (поддерживаем оба формата)
  const allBoxes = washInfo?.allBoxes || washInfo?.all_boxes || [];
  const washQueue = washInfo?.washQueue || washInfo?.wash_queue || { queue_size: 0, has_queue: false };
  const airDryQueue = washInfo?.airDryQueue || washInfo?.air_dry_queue || { queue_size: 0, has_queue: false };
  const vacuumQueue = washInfo?.vacuumQueue || washInfo?.vacuum_queue || { queue_size: 0, has_queue: false };
  
  // Используем userSession из washInfo
  const userSession = washInfo?.userSession || washInfo?.user_session;
  const payment = washInfo?.payment;
  
  // Используем хук для таймера
  const { timeLeft } = useTimer(userSession);
  
  // Проверяем, можно ли отменить сессию
  const canCancelSession = userSession && ['created', 'in_queue', 'assigned'].includes(userSession.status);
  
  // Проверяем, можно ли продлить сессию (всегда когда сессия активна и время не истекло)
  const canExtendSession = userSession && 
    userSession.status === 'active' && 
    timeLeft !== null && 
    timeLeft > 0; // Время еще не истекло
  
  // Проверяем, можно ли продлить сессию при неуспешной оплате продления
  const canRetryExtension = userSession && 
    userSession.status === 'active' && 
    userSession.requested_extension_time_minutes > 0 && // Запрошено продление
    payment && 
    (payment.status === 'failed' || payment.status === 'pending'); // Но оплата неуспешна
  
  // Проверяем, можно ли докупить химию (химия не была куплена или полностью использована)
  const canBuyChemistry = userSession && 
    userSession.status === 'active' && 
    userSession.service_type === 'wash' &&
    availableChemistryTimes.length > 0; // Добавляем проверку на наличие доступных вариантов
  
  // Получаем информацию о возврате
  const refundInfo = formatRefundInfo(payment);
  
  // Функция для расчета оставшегося времени химии в минутах
  const calculateRemainingChemistryTime = React.useCallback((session) => {
    if (!session) return 0;
    
    // Если химия не была включена, возвращаем всю купленную химию
    if (!session.was_chemistry_on || !session.chemistry_started_at) {
      return session.chemistry_time_minutes || 0;
    }
    
    // Если химия была выключена, значит использована полностью
    if (session.chemistry_ended_at) {
      return 0;
    }
    
    // Если химия активна, рассчитываем оставшееся время
    if (session.chemistry_started_at && !session.chemistry_ended_at) {
      const startTime = new Date(session.chemistry_started_at);
      const now = new Date();
      const timeLimit = (session.chemistry_time_minutes || 0) * 60 * 1000; // в миллисекундах
      const timePassed = now - startTime;
      const remainingMs = timeLimit - timePassed;
      
      if (remainingMs <= 0) {
        return 0;
      }
      
      // Конвертируем в минуты
      return Math.floor(remainingMs / (60 * 1000));
    }
    
    return 0;
  }, []);
  
  // Функция для загрузки платежей сессии
  const loadSessionPayments = async () => {
    if (!userSession || !userSession.id) return;
    
    try {
      setLoadingPayments(true);
      const payments = await ApiService.getSessionPayments(userSession.id);
      setSessionPayments(payments);
    } catch (error) {
      console.error('Ошибка при загрузке платежей сессии:', error);
    } finally {
      setLoadingPayments(false);
    }
  };
  
  // Загружаем платежи при изменении сессии
  useEffect(() => {
    if (userSession && userSession.id) {
      loadSessionPayments();
    }
  }, [userSession?.id]);
  
  // Предзагружаем доступные опции химии для докупки
  useEffect(() => {
    // Не пересчитываем, если открыт модал продления или докупки
    if (showExtendModal || showBuyChemistryModal) {
      return;
    }
    
    if (userSession && userSession.status === 'active' && userSession.service_type === 'wash') {
      // Загружаем настройки один раз
      if (allChemistryTimesFromSettings.length === 0) {
        const loadSettings = async () => {
          try {
            const response = await ApiService.getAvailableChemistryTimes(userSession.service_type);
            if (response && response.available_chemistry_times) {
              setAllChemistryTimesFromSettings(response.available_chemistry_times);
              
              // Вычисляем доступные опции для докупки (без продления)
              const remainingWashMinutes = timeLeft ? Math.floor(timeLeft / 60) : 0;
              const remainingChemistryMinutes = calculateRemainingChemistryTime(userSession);
              const availableTime = remainingWashMinutes - remainingChemistryMinutes;
              
              const filteredTimes = response.available_chemistry_times.filter(time => time <= availableTime);
              setAvailableChemistryTimes(filteredTimes);
            }
          } catch (err) {
            console.error('Ошибка при загрузке доступного времени химии:', err);
          }
        };
        loadSettings();
      } else {
        // Если настройки уже загружены, просто пересчитываем доступные опции для докупки
        const remainingWashMinutes = timeLeft ? Math.floor(timeLeft / 60) : 0;
        const remainingChemistryMinutes = calculateRemainingChemistryTime(userSession);
        const availableTime = remainingWashMinutes - remainingChemistryMinutes;
        
        const filteredTimes = allChemistryTimesFromSettings.filter(time => time <= availableTime);
        setAvailableChemistryTimes(filteredTimes);
      }
    }
  }, [userSession?.id, userSession?.status, userSession?.service_type, userSession?.was_chemistry_on, userSession?.chemistry_started_at, userSession?.chemistry_ended_at, timeLeft, allChemistryTimesFromSettings, calculateRemainingChemistryTime, showExtendModal, showBuyChemistryModal]);
  
  // Функция для перехода на страницу сессии
  const handleViewSessionDetails = () => {
    try {
      if (userSession && userSession.id) {
        navigate(`/telegram/session/${userSession.id}`);
      }
    } catch (error) {
      alert('Ошибка при переходе к деталям сессии: ' + error.message);
    }
  };

  // Обработчик нажатия на кнопку "Записаться на мойку"
  const handleCreateSessionClick = () => {
    try {
      navigate('/telegram/booking');
    } catch (error) {
      alert('Ошибка при переходе на страницу записи: ' + error.message);
    }
  };


  // Обработчик отмены сессии
  const handleCancelSession = async () => {
    if (!userSession || !user) return;
    
    const confirmMessage = refundInfo.hasRefund 
      ? `Вы уверены, что хотите отменить сессию? Деньги в размере ${formatAmountWithRefund(payment)} будут возвращены на карту.`
      : 'Вы уверены, что хотите отменить сессию?';
    
    if (!window.confirm(confirmMessage)) return;
    
    try {
      setIsCanceling(true);
      await onCancelSession(userSession.id, user.id);
    } catch (error) {
      alert('Ошибка при отмене сессии: ' + error.message);
    } finally {
      setIsCanceling(false);
    }
  };

  // Функция для загрузки доступного времени мойки
  const fetchAvailableRentalTimes = async (serviceType) => {
    try {
      setLoadingRentalTimes(true);
      
      // Если настройки еще не загружены, загружаем их
      if (allRentalTimesFromSettings.length === 0) {
        const response = await ApiService.getAvailableRentalTimes(serviceType);
        if (response && response.available_times) {
          setAllRentalTimesFromSettings(response.available_times);
          setAvailableRentalTimes(response.available_times);
          // Устанавливаем первое значение как выбранное по умолчанию
          if (response.available_times.length > 0) {
            setSelectedExtensionTime(response.available_times[0]);
          }
        }
      } else {
        // Используем кэшированные настройки
        setAvailableRentalTimes(allRentalTimesFromSettings);
        if (allRentalTimesFromSettings.length > 0 && !selectedExtensionTime) {
          setSelectedExtensionTime(allRentalTimesFromSettings[0]);
        }
      }
    } catch (err) {
      alert('Ошибка при загрузке доступного времени мойки: ' + err.message);
    } finally {
      setLoadingRentalTimes(false);
    }
  };

  // Функция для загрузки доступного времени химии
  const fetchAvailableChemistryTimes = async (serviceType, forExtension = false, extensionTime = 0) => {
    try {
      setLoadingChemistryTimes(true);
      
      // Если настройки еще не загружены, загружаем их
      if (allChemistryTimesFromSettings.length === 0) {
        const response = await ApiService.getAvailableChemistryTimes(serviceType);
        if (response && response.available_chemistry_times) {
          setAllChemistryTimesFromSettings(response.available_chemistry_times);
        }
      }
      
      // Используем кэшированные настройки или загружаем
      const allChemistryTimes = allChemistryTimesFromSettings.length > 0 
        ? allChemistryTimesFromSettings 
        : [];
      
      if (allChemistryTimes.length > 0) {
        // Вычисляем доступное время для химии
        const remainingWashMinutes = timeLeft ? Math.floor(timeLeft / 60) : 0;
        const remainingChemistryMinutes = calculateRemainingChemistryTime(userSession);
        
        // Для продления добавляем время продления к оставшемуся времени мойки
        const availableTime = forExtension 
          ? remainingWashMinutes + extensionTime - remainingChemistryMinutes
          : remainingWashMinutes - remainingChemistryMinutes;
        
        // Фильтруем только те опции, которые меньше или равны доступному времени
        const filteredTimes = allChemistryTimes.filter(time => time <= availableTime);
        
        setAvailableChemistryTimes(filteredTimes);
      }
    } catch (err) {
      console.error('Ошибка при загрузке доступного времени химии:', err);
    } finally {
      setLoadingChemistryTimes(false);
    }
  };

  // Функция для продления сессии с оплатой
  const handleExtendSession = async () => {
    if (!selectedExtensionTime) {
      alert('Выберите время продления');
      return;
    }

    try {
      setActionLoading(true);
      
      // Вызываем API для продления сессии с оплатой
      const response = await ApiService.extendSessionWithPayment(userSession.id, selectedExtensionTime, selectedChemistryTime);
      
      if (response && response.payment) {
        // Перенаправляем на страницу оплаты
        navigate('/telegram/payment', { 
          state: { 
            session: response.session,
            payment: response.payment,
            paymentType: 'extension',
            sessionId: userSession.id
          } 
        });
      }
    } catch (err) {
      alert('Ошибка при создании платежа продления: ' + err.message);
    } finally {
      setActionLoading(false);
    }
  };

  // Функция для докупки химии
  const handleBuyChemistry = async () => {
    if (!selectedChemistryTime) {
      alert('Выберите время химии');
      return;
    }

    try {
      setActionLoading(true);
      
      // Вызываем API для докупки химии (ExtensionTimeMinutes = 0, ExtensionChemistryTimeMinutes = selectedChemistryTime)
      const response = await ApiService.extendSessionWithPayment(userSession.id, 0, selectedChemistryTime);
      
      if (response && response.payment) {
        // Перенаправляем на страницу оплаты
        navigate('/telegram/payment', { 
          state: { 
            session: response.session,
            payment: response.payment,
            paymentType: 'extension',
            sessionId: userSession.id
          } 
        });
      }
    } catch (err) {
      alert('Ошибка при создании платежа докупки химии: ' + err.message);
    } finally {
      setActionLoading(false);
    }
  };

  // Функция для открытия модального окна продления
  const openExtendModal = () => {
    if (userSession && userSession.service_type) {
      // Сбрасываем предыдущие значения при повторном продлении
      setSelectedExtensionTime(null);
      setSelectedChemistryTime(null);
      
      fetchAvailableRentalTimes(userSession.service_type);
      // Загружаем время химии только для мойки с химией
      if (userSession.service_type === 'wash' && userSession.with_chemistry) {
        fetchAvailableChemistryTimes(userSession.service_type, true, 0);
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
    if (userSession && userSession.service_type === 'wash') {
      // Сбрасываем предыдущие значения
      setSelectedChemistryTime(null);
      
      // Загружаем доступное время химии
      fetchAvailableChemistryTimes(userSession.service_type);
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
    // При изменении времени продления пересчитываем доступное время химии
    if (userSession && userSession.service_type === 'wash' && userSession.with_chemistry) {
      fetchAvailableChemistryTimes(userSession.service_type, true, time);
    }
  };

  // Функция для выбора времени химии
  const handleChemistryTimeSelect = (time) => {
    setSelectedChemistryTime(time);
  };

  // Функция для завершения сессии
  const handleCompleteSession = async () => {
    try {
      setActionLoading(true);
      
      // Вызываем API для завершения сессии
      const response = await ApiService.completeSession(userSession.id);
      
      if (response && response.session) {
        // Немедленно обновляем данные сессии для мгновенного отображения изменений
        try {
          const updatedSessionData = await ApiService.getSessionById(userSession.id);
          if (updatedSessionData && updatedSessionData.session) {
            // Обновляем данные через callback, если он передан
            if (onCompleteSession) {
              onCompleteSession(updatedSessionData.session, updatedSessionData.payment);
            }
          }
        } catch (refreshError) {
          console.error('Ошибка при обновлении данных сессии:', refreshError);
          // Не показываем ошибку пользователю, поллинг все равно обновит данные
        }
      }
    } catch (err) {
      alert('Ошибка при завершении сессии: ' + err.message);
    } finally {
      setActionLoading(false);
    }
  };

  // Функция для запуска сессии
  const handleStartSession = async () => {
    try {
      setActionLoading(true);
      
      // Вызываем API для обновления статуса сессии
      const response = await ApiService.startSession(userSession.id);
      
      if (response && response.session) {
        // Поллинг автоматически обновит данные через несколько секунд
        // Не нужно перезагружать страницу
      }
    } catch (err) {
      alert('Ошибка при запуске сессии: ' + err.message);
    } finally {
      setActionLoading(false);
    }
  };

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  return (
    <div className={styles.container}>
      {/* Кнопка записи на мойку - показывается только если нет сессии */}
      {!userSession && (
        <section className={styles.section}>
          <Card theme={theme} style={{ padding: '24px' }}>
            <Button 
              theme={theme} 
              onClick={handleCreateSessionClick}
              className={styles.createSessionButton}
              style={{ 
                width: '100%',
                padding: '20px',
                fontSize: '18px',
                fontWeight: 'bold',
                minHeight: '60px',
                backgroundColor: '#4CAF50',
                color: 'white',
                border: 'none',
                borderRadius: '12px',
                boxShadow: '0 4px 8px rgba(0, 0, 0, 0.1)',
                transition: 'all 0.3s ease'
              }}
            >
              НАЖМИ, чтобы помыть машину/записаться в очередь
            </Button>
          </Card>
        </section>
      )}

      {/* Информация о сессии пользователя - если есть сессия */}
      {userSession && (
        <section className={styles.section}>
          <Card theme={theme}>
            <StatusBadge status={userSession.status} theme={theme} />
            
            {/* Номер бокса с цветным фоном */}
            {(userSession.boxId || userSession.box_id || userSession.boxNumber || userSession.box_number) && (
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
                  Бокс #{
                    userSession.boxNumber || userSession.box_number || 
                    allBoxes.find(box => box.id === (userSession.boxId || userSession.box_id))?.number || 
                    'Неизвестный бокс'
                  }
                  {boxChanged && <span style={{ color: '#856404', fontSize: '12px', marginLeft: '8px' }}>🔄 Обновлено!</span>}
                </div>
              </div>
            )}
            
            {/* Таймеры для активной сессии */}
            {userSession.status === 'active' && timeLeft !== null && (
              <>
                <p className={`${styles.sessionInfo} ${themeClass}`} style={{ marginTop: '12px', fontWeight: 'bold' }}>
                  Оставшееся время мойки:
                </p>
                <Timer seconds={timeLeft} theme={theme} />
                
                {/* Статус и таймер химии (если была включена) */}
                {(userSession.withChemistry || userSession.with_chemistry) && 
                 (userSession.wasChemistryOn || userSession.was_chemistry_on) && (
                  <ChemistryStatus session={userSession} />
                )}
              </>
            )}
            
            {/* Таймеры для назначенной сессии */}
            {userSession.status === 'assigned' && timeLeft !== null && (
              <>
                <p className={`${styles.sessionInfo} ${themeClass}`} style={{ marginTop: '12px', fontWeight: 'bold' }}>
                  Время до старта мойки:
                </p>
                <Timer seconds={timeLeft} theme={theme} />
              </>
            )}
            
            {/* Кнопки под таймерами */}
            <div style={{ marginTop: '12px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
              {/* Кнопка включения химии для активной сессии */}
              {userSession.status === 'active' && 
               userSession.with_chemistry && 
               userSession.chemistry_time_minutes > 0 && 
               !userSession.was_chemistry_on && (
                <div style={{ marginTop: '12px' }}>
                  <p style={{ fontSize: '12px', color: '#666', marginBottom: '8px' }}>
                    Оплачено: {userSession.chemistry_time_minutes} мин. химии
                  </p>
                  <Button 
                    theme={theme} 
                  onClick={async () => {
                    try {
                      await ApiService.enableChemistry(userSession.id);
                      // Поллинг автоматически обновит данные через несколько секунд
                      // Не нужно перезагружать страницу
                    } catch (error) {
                      console.error('Ошибка включения химии:', error);
                      alert('Ошибка включения химии: ' + (error.response?.data?.error || error.message));
                    }
                  }}
                    style={{ 
                      backgroundColor: '#4CAF50',
                      color: 'white',
                      width: '100%'
                    }}
                  >
                    🧪 Включить химию
                  </Button>
                </div>
              )}
              
              {/* Кнопки для активной сессии */}
              {userSession.status === 'active' && (
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
              {userSession.status === 'assigned' && userSession.box_id && (
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
                    color: 'white',
                    width: '100%'
                  }}
                >
                  {isCanceling ? 'Отмена...' : 'Отменить сессию'}
                </Button>
              )}
            </div>
            
            {/* Информация о платеже */}
            {payment && (
              <div style={{ 
                marginTop: '12px',
                padding: '8px',
                backgroundColor: '#E8F5E8',
                borderRadius: '4px',
                fontSize: '12px'
              }}>
                <p style={{ margin: '0 0 4px 0', color: '#2E7D32', fontWeight: 'bold' }}>
                  💰 Стоимость: {loadingPayments ? 'Загрузка...' : sessionPayments ? formatSessionTotalCost(sessionPayments) : formatAmountWithRefund(payment)}
                </p>
                {refundInfo.hasRefund && (
                  <>
                    <p style={{ margin: '0 0 4px 0', color: '#1976D2', fontWeight: 'bold' }}>
                      💸 Возвращено: {formatAmount(refundInfo.refundedAmount)}
                      {refundInfo.refundType === 'partial' && ` (частично)`}
                      {refundInfo.refundType === 'full' && ` (полностью)`}
                    </p>
                  </>
                )}
                <p style={{ margin: '0', color: '#2E7D32' }}>
                  {payment.status === 'succeeded' ? '✅ Оплачено' :
                   payment.status === 'pending' ? '⏳ Ожидает оплаты' :
                   payment.status === 'failed' ? '❌ Ошибка оплаты' :
                   payment.status === 'refunded' ? '💸 Полностью возвращено' : '❓ Неизвестный статус'}
                </p>
                {refundInfo.hasRefund && (
                  <p style={{ margin: '4px 0 0 0', color: '#1976D2', fontWeight: 'bold' }}>
                    💰 Итого: {formatAmount(refundInfo.finalAmount)}
                  </p>
                )}
              </div>
            )}
            
            {/* Показываем информацию для созданной сессии (ожидание оплаты) */}
            {userSession.status === 'created' && (
              <div className={`${styles.sessionInfo} ${themeClass}`} style={{ 
                marginTop: '12px',
                padding: '12px',
                backgroundColor: '#FFF3E0',
                borderRadius: '8px',
                border: '1px solid #FFB74D'
              }}>
                <p style={{ margin: '0 0 12px 0', fontSize: '14px' }}>
                  Сессия создана, но оплата еще не произведена. 
                  После оплаты сессия будет добавлена в очередь.
                </p>
                
                {/* Кнопка оплаты */}
                <Button 
                  theme={theme} 
                  onClick={async () => {
                    try {
                      // Запрашиваем последний платеж по сессии
                      const response = await ApiService.getUserSessionForPayment(userSession.user_id);
                      
                      navigate('/telegram/payment', {
                        state: {
                          session: response.session,
                          payment: response.payment,
                          sessionId: userSession.id
                        }
                      });
                    } catch (error) {
                      console.error('Ошибка получения платежа:', error);
                      alert('Ошибка получения платежа: ' + error.message);
                    }
                  }}
                  style={{ 
                    marginTop: '8px',
                    backgroundColor: '#FF9800',
                    color: 'white',
                    width: '100%'
                  }}
                >
                  💳 Оплатить
                </Button>
              </div>
            )}
            
            {/* Показываем информацию для сессии с ошибкой оплаты */}
            {userSession.status === 'payment_failed' && (
              <div className={`${styles.sessionInfo} ${themeClass}`} style={{ 
                marginTop: '12px',
                padding: '12px',
                backgroundColor: '#FFEBEE',
                borderRadius: '8px',
                border: '1px solid #E57373'
              }}>
                <p style={{ margin: '0 0 8px 0', fontWeight: 'bold', color: '#C62828' }}>
                  ❌ Ошибка оплаты
                </p>
                <p style={{ margin: '0 0 12px 0', fontSize: '14px' }}>
                  Произошла ошибка при оплате. 
                  Попробуйте создать новую сессию или повторить оплату.
                </p>
                
                {/* Показываем информацию о платеже, если есть */}
                {payment && (
                  <div style={{ 
                    marginBottom: '12px',
                    padding: '8px',
                    backgroundColor: '#FFCDD2',
                    borderRadius: '4px',
                    fontSize: '12px'
                  }}>
                    <p style={{ margin: '0 0 4px 0', color: '#C62828', fontWeight: 'bold' }}>
                      💰 Стоимость: {loadingPayments ? 'Загрузка...' : sessionPayments ? formatSessionTotalCost(sessionPayments) : formatAmountWithRefund(payment)}
                    </p>
                    {refundInfo.hasRefund && (
                      <p style={{ margin: '0 0 4px 0', color: '#1976D2', fontWeight: 'bold' }}>
                        💸 Возвращено: {formatAmount(refundInfo.refundedAmount)}
                      </p>
                    )}
                    <p style={{ margin: '0', color: '#C62828' }}>
                      ❌ Статус: {getPaymentStatusText(payment.status)}
                    </p>
                    {refundInfo.hasRefund && (
                      <p style={{ margin: '4px 0 0 0', color: '#1976D2', fontWeight: 'bold' }}>
                        💰 Итого: {formatAmount(refundInfo.finalAmount)}
                      </p>
                    )}
                  </div>
                )}
              </div>
            )}
            
            {/* Показываем информацию, если таймер не отображается */}
            {userSession.status === 'assigned' && timeLeft === null && (
              <p className={`${styles.sessionInfo} ${themeClass}`} style={{ 
                color: '#C62828', 
                textAlign: 'center',
                fontSize: '12px'
              }}>
                ⚠️ Таймер не отображается (timeLeft = null)
              </p>
            )}
          </Card>
        </section>
      )}

      {/* Информация об очереди */}
      <section className={styles.section}>
        <h2 className={`${styles.title} ${themeClass}`}>Статус автомойки</h2>
        <Card theme={theme}>
          {/* Информация о разных типах очередей */}
          <div className={styles.queueTypesContainer}>
            <div className={styles.queueTypeItem}>
              <h4 className={`${styles.queueTypeTitle} ${themeClass}`}>Мойка</h4>
              <StatusBadge 
                status={washQueue.has_queue ? 'busy' : 'free'} 
                theme={theme}
                text={formatQueueText(washQueue)}
              />
            </div>
            <div className={styles.queueTypeItem}>
              <h4 className={`${styles.queueTypeTitle} ${themeClass}`}>Пылесос</h4>
              <StatusBadge 
                status={vacuumQueue.has_queue ? 'busy' : 'free'} 
                theme={theme}
                text={formatQueueText(vacuumQueue)}
              />
            </div>
            <div className={styles.queueTypeItem}>
              <h4 className={`${styles.queueTypeTitle} ${themeClass}`}>Воздух</h4>
              <StatusBadge 
                status={airDryQueue.has_queue ? 'busy' : 'free'} 
                theme={theme}
                text={formatQueueText(airDryQueue)}
              />
            </div>
          </div>
        </Card>
      </section>

      {/* Кнопка для просмотра истории сессий */}
      <section className={styles.section}>
        <h2 className={`${styles.title} ${themeClass}`}>История моек</h2>
        <Card theme={theme}>
          <Button 
            theme={theme} 
            onClick={onViewHistory}
            className={styles.historyButton}
          >
            Посмотреть историю моек
          </Button>
        </Card>
      </section>

      {/* Модальное окно для продления сессии */}
      {showExtendModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '20px'
        }}>
          <Card theme={theme} style={{ maxWidth: '400px', width: '100%', maxHeight: '80vh', overflow: 'auto' }}>
            <h3 style={{ margin: '0 0 16px 0', fontSize: '18px' }}>Продление сессии</h3>
            
            {loadingRentalTimes ? (
              <p>Загрузка доступного времени...</p>
            ) : (
              <>
                <p style={{ margin: '0 0 12px 0' }}>Выберите время продления:</p>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(80px, 1fr))', gap: '8px', marginBottom: '16px' }}>
                  {availableRentalTimes.map((time) => (
                    <div 
                      key={time} 
                      style={{
                        padding: '12px',
                        border: selectedExtensionTime === time ? '2px solid #2196F3' : '1px solid #ddd',
                        borderRadius: '8px',
                        textAlign: 'center',
                        cursor: 'pointer',
                        backgroundColor: selectedExtensionTime === time ? '#E3F2FD' : 'white'
                      }}
                      onClick={() => handleExtensionTimeSelect(time)}
                    >
                      <span style={{ fontSize: '16px', fontWeight: 'bold' }}>{time}</span>
                      <span style={{ fontSize: '12px', color: '#666' }}> мин</span>
                    </div>
                  ))}
                </div>
                
                {/* Выбор времени химии для мойки с химией - показываем только если есть доступные опции */}
                {userSession && userSession.service_type === 'wash' && userSession.with_chemistry && !loadingChemistryTimes && availableChemistryTimes.length > 0 && (
                  <>
                    <p style={{ margin: '20px 0 12px 0' }}>
                      Дополнительное время химии (опционально):
                    </p>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(80px, 1fr))', gap: '8px', marginBottom: '16px' }}>
                      {availableChemistryTimes.map((time) => (
                        <div 
                          key={time} 
                          style={{
                            padding: '12px',
                            border: selectedChemistryTime === time ? '2px solid #2196F3' : '1px solid #ddd',
                            borderRadius: '8px',
                            textAlign: 'center',
                            cursor: 'pointer',
                            backgroundColor: selectedChemistryTime === time ? '#E3F2FD' : 'white'
                          }}
                          onClick={() => handleChemistryTimeSelect(time)}
                        >
                          <span style={{ fontSize: '16px', fontWeight: 'bold' }}>{time}</span>
                          <span style={{ fontSize: '12px', color: '#666' }}> мин</span>
                        </div>
                      ))}
                    </div>
                  </>
                )}
                
                {/* Показываем текст загрузки только если загружаем и нет доступных опций */}
                {userSession && userSession.service_type === 'wash' && userSession.with_chemistry && loadingChemistryTimes && (
                  <p style={{ margin: '20px 0 12px 0' }}>Загрузка времени химии...</p>
                )}
                
                <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
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
          </Card>
        </div>
      )}
      
      {/* Модальное окно для докупки химии */}
      {showBuyChemistryModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '20px'
        }}>
          <Card theme={theme} style={{ maxWidth: '400px', width: '100%', maxHeight: '80vh', overflow: 'auto' }}>
            <h3 style={{ margin: '0 0 16px 0', fontSize: '18px' }}>Докупка химии</h3>
            
            {loadingChemistryTimes ? (
              <p>Загрузка доступного времени химии...</p>
            ) : (
              <>
                <p style={{ margin: '0 0 12px 0' }}>Выберите время химии:</p>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(80px, 1fr))', gap: '8px', marginBottom: '16px' }}>
                  {availableChemistryTimes.map((time) => (
                    <div 
                      key={time} 
                      style={{
                        padding: '12px',
                        border: selectedChemistryTime === time ? '2px solid #2196F3' : '1px solid #ddd',
                        borderRadius: '8px',
                        textAlign: 'center',
                        cursor: 'pointer',
                        backgroundColor: selectedChemistryTime === time ? '#E3F2FD' : 'white'
                      }}
                      onClick={() => handleChemistryTimeSelect(time)}
                    >
                      <span style={{ fontSize: '16px', fontWeight: 'bold' }}>{time}</span>
                      <span style={{ fontSize: '12px', color: '#666' }}> мин</span>
                    </div>
                  ))}
                </div>
                
                <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
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
          </Card>
        </div>
      )}
    </div>
  );
};

export default WashInfo;
