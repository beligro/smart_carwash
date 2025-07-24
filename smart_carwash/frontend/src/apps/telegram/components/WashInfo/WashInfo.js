import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './WashInfo.module.css';
import { Card, Button, StatusBadge, Timer } from '../../../../shared/components/UI';
import ServiceSelector from '../ServiceSelector';
import { formatDate } from '../../../../shared/utils/formatters';
import { getSessionStatusDescription, getServiceTypeDescription, formatRefundInfo, formatAmount, getPaymentStatusText, getPaymentStatusColor } from '../../../../shared/utils/statusHelpers';
import useTimer from '../../../../shared/hooks/useTimer';

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
const WashInfo = ({ washInfo, theme = 'light', onCreateSession, onViewHistory, onCancelSession, user }) => {
  const navigate = useNavigate();
  const [showServiceSelector, setShowServiceSelector] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  
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
      setShowServiceSelector(true);
    } catch (error) {
      alert('Ошибка при открытии выбора услуг: ' + error.message);
    }
  };

  // Обработчик выбора услуги
  const handleServiceSelect = (serviceData) => {
    try {
      setShowServiceSelector(false);
      // Используем новый метод создания сессии с платежом
      onCreateSession(serviceData);
    } catch (error) {
      alert('Ошибка при выборе услуги: ' + error.message);
    }
  };

  // Обработчик отмены сессии
  const handleCancelSession = async () => {
    if (!userSession || !user) return;
    
    const confirmMessage = refundInfo.hasRefund 
      ? `Вы уверены, что хотите отменить сессию? Деньги в размере ${formatAmount(payment.amount)} будут возвращены на карту.`
      : 'Вы уверены, что хотите отменить сессию?';
    
    if (!window.confirm(confirmMessage)) return;
    
    try {
      setIsCanceling(true);
      await onCancelSession(userSession.id, user.id);
      alert('Сессия успешно отменена' + (refundInfo.hasRefund ? '. Деньги будут возвращены на карту.' : ''));
    } catch (error) {
      alert('Ошибка при отмене сессии: ' + error.message);
    } finally {
      setIsCanceling(false);
    }
  };

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  // Если открыт выбор услуг, показываем только его
  if (showServiceSelector) {
    return (
      <div className={styles.container}>
        <ServiceSelector 
          onSelect={handleServiceSelect} 
          theme={theme} 
          user={user}
        />
        <div className={styles.buttonContainer}>
          <Button 
            theme={theme} 
            onClick={() => setShowServiceSelector(false)}
            className={styles.cancelButton}
          >
            Отмена
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
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
                text={washQueue.has_queue ? `В очереди: ${washQueue.queue_size}` : 'Нет очереди'}
              />
            </div>
            <div className={styles.queueTypeItem}>
              <h4 className={`${styles.queueTypeTitle} ${themeClass}`}>Обдув</h4>
              <StatusBadge 
                status={airDryQueue.has_queue ? 'busy' : 'free'} 
                theme={theme}
                text={airDryQueue.has_queue ? `В очереди: ${airDryQueue.queue_size}` : 'Нет очереди'}
              />
            </div>
            <div className={styles.queueTypeItem}>
              <h4 className={`${styles.queueTypeTitle} ${themeClass}`}>Пылесос</h4>
              <StatusBadge 
                status={vacuumQueue.has_queue ? 'busy' : 'free'} 
                theme={theme}
                text={vacuumQueue.has_queue ? `В очереди: ${vacuumQueue.queue_size}` : 'Нет очереди'}
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

      {/* Информация о сессии пользователя */}
      <section className={styles.section}>
        <h2 className={`${styles.title} ${themeClass}`}>Ваша запись</h2>
        <Card theme={theme}>
          {userSession ? (
            <>
              <StatusBadge status={userSession.status} theme={theme} />
              <p className={`${styles.sessionInfo} ${themeClass}`}>
                Создана: {formatDate(userSession.createdAt || userSession.created_at)}
              </p>
              <p className={`${styles.sessionInfo} ${themeClass}`}>
                Услуга: {getServiceTypeDescription(userSession.serviceType || userSession.service_type)}
                {(userSession.withChemistry || userSession.with_chemistry) && ' (с химией)'}
              </p>
              <p className={`${styles.sessionInfo} ${themeClass}`}>
                Номер машины: {userSession.carNumber || userSession.car_number || 'Не указан'}
              </p>
              {(userSession.boxId || userSession.box_id || userSession.boxNumber || userSession.box_number) && (
                <p className={`${styles.sessionInfo} ${themeClass}`}>
                  Назначен бокс: #{
                    // Используем номер бокса из сессии, если он есть
                    userSession.boxNumber || userSession.box_number || 
                    // Иначе находим номер бокса по его ID
                    allBoxes.find(box => box.id === (userSession.boxId || userSession.box_id))?.number || 
                    'Неизвестный бокс'
                  }
                </p>
              )}
              <div className={`${styles.statusIndicator} ${themeClass}`}>
                <div className={`${styles.statusDot} ${styles[userSession.status]}`}></div>
                <span className={`${styles.statusText} ${themeClass}`}>
                  {getSessionStatusDescription(userSession.status)}
                </span>
              </div>
              
              {/* Информация о платеже для активной сессии */}
              {payment && (
                <div style={{ 
                  marginTop: '12px',
                  padding: '8px',
                  backgroundColor: '#E8F5E8',
                  borderRadius: '4px',
                  fontSize: '12px'
                }}>
                  <p style={{ margin: '0 0 4px 0', color: '#2E7D32', fontWeight: 'bold' }}>
                    💰 Стоимость: {formatAmount(payment.amount)}
                  </p>
                  {refundInfo.hasRefund && (
                    <>
                      <p style={{ margin: '0 0 4px 0', color: '#1976D2', fontWeight: 'bold' }}>
                        💸 Возвращено: {formatAmount(refundInfo.refundedAmount)}
                        {refundInfo.refundType === 'partial' && ` (частично)`}
                        {refundInfo.refundType === 'full' && ` (полностью)`}
                      </p>
                      {refundInfo.refundType === 'partial' && (
                        <p style={{ margin: '0 0 4px 0', color: '#FF9800', fontWeight: 'bold' }}>
                          💰 Осталось к возврату: {formatAmount(refundInfo.remainingAmount)}
                        </p>
                      )}
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
              
              {/* Отображаем таймер для активной сессии */}
              {userSession.status === 'active' && timeLeft !== null && (
                <>
                  <p className={`${styles.sessionInfo} ${themeClass}`} style={{ marginTop: '12px', fontWeight: 'bold' }}>
                    Оставшееся время мойки:
                  </p>
                  <Timer seconds={timeLeft} theme={theme} />
                </>
              )}
              
              {/* Отображаем таймер для назначенной сессии */}
              {userSession.status === 'assigned' && timeLeft !== null && (
                <>
                  <p className={`${styles.sessionInfo} ${themeClass}`} style={{ marginTop: '12px', fontWeight: 'bold' }}>
                    Время до истечения резерва:
                  </p>
                  <Timer seconds={timeLeft} theme={theme} />
                  <p className={`${styles.sessionInfo} ${themeClass}`} style={{ 
                    color: timeLeft <= 60 ? '#C62828' : 'inherit', 
                    textAlign: 'center' 
                  }}>
                    Начните мойку до истечения времени, иначе резерв будет снят
                  </p>
                </>
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
                    onClick={() => {
                      // Переходим на страницу оплаты с данными сессии
                      navigate('/telegram/payment', {
                        state: {
                          session: userSession,
                          payment: payment || null
                        }
                      });
                    }}
                    style={{ 
                      marginTop: '8px',
                      backgroundColor: '#FF9800',
                      color: 'white'
                    }}
                  >
                    💳 Оплатить
                  </Button>
                </div>
              )}
              
              {/* Показываем информацию для сессии в очереди */}
              {userSession.status === 'in_queue' && (
                <div className={`${styles.sessionInfo} ${themeClass}`} style={{ 
                  marginTop: '12px',
                  padding: '12px',
                  backgroundColor: '#E8F5E8',
                  borderRadius: '8px',
                  border: '1px solid #81C784'
                }}>
                  <p style={{ margin: '0 0 4px 0', color: '#2E7D32', fontWeight: 'bold' }}>
                    💰 Стоимость: {formatAmount(payment.amount)}
                  </p>
                  <p style={{ margin: '0 0 8px 0', fontWeight: 'bold', color: '#2E7D32' }}>
                    ✅ Оплачено, в очереди
                  </p>
                  <p style={{ margin: '0 0 12px 0', fontSize: '14px' }}>
                    Сессия оплачена и добавлена в очередь. 
                    Ожидайте назначения свободного бокса.
                  </p>
                  
                  {/* Показываем информацию о платеже, если есть */}
                  {payment && (
                    <div style={{ 
                      marginTop: '8px',
                      padding: '8px',
                      backgroundColor: '#F1F8E9',
                      borderRadius: '4px',
                      fontSize: '12px'
                    }}>
                      <p style={{ margin: '0 0 4px 0', color: '#2E7D32', fontWeight: 'bold' }}>
                        💰 Стоимость: {formatAmount(payment.amount)}
                      </p>
                      {refundInfo.hasRefund && (
                        <p style={{ margin: '0 0 4px 0', color: '#1976D2', fontWeight: 'bold' }}>
                          💸 Возвращено: {formatAmount(refundInfo.refundedAmount)}
                        </p>
                      )}
                      <p style={{ margin: '0', color: '#2E7D32' }}>
                        ✅ Статус: {getPaymentStatusText(payment.status)}
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
                        💰 Стоимость: {formatAmount(payment.amount)}
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
                  
                  {/* Кнопка повторной оплаты */}
                  <Button 
                    theme={theme} 
                    onClick={() => {
                      navigate('/telegram/payment', {
                        state: {
                          session: userSession,
                          payment: payment || null
                        }
                      });
                    }}
                    style={{ 
                      marginTop: '8px',
                      backgroundColor: '#F44336',
                      color: 'white'
                    }}
                  >
                    🔄 Повторить оплату
                  </Button>
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
              <Button 
                theme={theme} 
                onClick={handleViewSessionDetails}
              >
                Подробнее о сессии
              </Button>
              {canCancelSession && (
                <Button 
                  theme={theme} 
                  onClick={handleCancelSession}
                  disabled={isCanceling}
                  style={{ 
                    marginTop: '8px',
                    backgroundColor: '#F44336',
                    color: 'white'
                  }}
                >
                  {isCanceling ? 'Отмена...' : 'Отменить сессию'}
                </Button>
              )}
            </>
          ) : (
            <>
              <p className={`${styles.sessionInfo} ${themeClass}`}>
                У вас нет активной записи на мойку
              </p>
              <Button 
                theme={theme} 
                onClick={handleCreateSessionClick}
                className={styles.createSessionButton}
              >
                Записаться на мойку
              </Button>
            </>
          )}
        </Card>
      </section>
    </div>
  );
};

export default WashInfo;
