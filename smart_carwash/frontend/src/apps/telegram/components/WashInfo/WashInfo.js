import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './WashInfo.module.css';
import { Card, Button, StatusBadge, Timer } from '../../../../shared/components/UI';
import { formatDate } from '../../../../shared/utils/formatters';
import { getSessionStatusDescription, getServiceTypeDescription, formatRefundInfo, formatAmount, formatAmountWithRefund, getPaymentStatusText, getPaymentStatusColor, formatSessionTotalCost, formatSessionDetailedCost } from '../../../../shared/utils/statusHelpers';
import useTimer from '../../../../shared/hooks/useTimer';
import ApiService from '../../../../shared/services/ApiService';

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
const WashInfo = ({ washInfo, theme = 'light', onCreateSession, onViewHistory, onCancelSession, onChemistryEnabled, user }) => {
  const navigate = useNavigate();
  const [isCanceling, setIsCanceling] = useState(false);
  const [sessionPayments, setSessionPayments] = useState(null);
  const [loadingPayments, setLoadingPayments] = useState(false);
  const [boxChanged, setBoxChanged] = useState(false);
  
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
  
  // Получаем информацию о возврате
  const refundInfo = formatRefundInfo(payment);
  
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

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  return (
    <div className={styles.container}>
      {/* Кнопка записи на мойку - показывается только если нет сессии */}
      {!userSession && (
        <section className={styles.section}>
          <Card theme={theme}>
            <Button 
              theme={theme} 
              onClick={handleCreateSessionClick}
              className={styles.createSessionButton}
              style={{ width: '100%' }}
            >
              Помыть машину/записаться в очередь
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
              <Button 
                theme={theme} 
                onClick={handleViewSessionDetails}
                style={{ width: '100%' }}
              >
                Подробнее о сессии
              </Button>
              {canCancelSession && (
                <Button 
                  theme={theme} 
                  onClick={handleCancelSession}
                  disabled={isCanceling}
                  style={{ 
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
    </div>
  );
};

export default WashInfo;
