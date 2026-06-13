import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './SessionHistory.module.css';
import { Card, Button, StatusBadge } from '../../../../shared/components/UI';
import { formatDate } from '../../../../shared/utils/formatters';
import { getSessionStatusDescription, formatSessionTotalCost } from '../../../../shared/utils/statusHelpers';
import WebApiService from '../../../../shared/services/WebApiService';

/**
 * Компонент SessionHistory - отображает историю сессий пользователя
 * @param {Object} props - Свойства компонента
 * @param {Object} props.user - Информация о пользователе
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 */
const SessionHistory = ({ user, theme = 'light', basePath: basePathProp }) => {
  const navigate = useNavigate();
  const pathBase = basePathProp ?? '/web';
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const limit = 5; // Количество сессий на странице

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  // Загрузка истории сессий
  useEffect(() => {
    if (!user || !user.id) return;

    const fetchSessionHistory = async () => {
      try {
        setLoading(true);
        const offset = page * limit;
        const response = await WebApiService.getUserSessionHistory(limit, offset);
        
        if (response && response.sessions) {
          if (page === 0) {
            setSessions(response.sessions);
          } else {
            setSessions(prevSessions => [...prevSessions, ...response.sessions]);
          }
          
          // Проверяем, есть ли еще сессии для загрузки
          setHasMore(response.sessions.length === limit);
        } else {
          setSessions([]);
          setHasMore(false);
        }
        
        setError(null);
      } catch (err) {
        alert('Ошибка при загрузке истории сессий: ' + err.message);
        setError('Не удалось загрузить историю сессий');
      } finally {
        setLoading(false);
      }
    };

    fetchSessionHistory();
  }, [user, page, limit]);

  // Функция для загрузки следующей страницы
  const loadMore = () => {
    setPage(prevPage => prevPage + 1);
  };

  // Функция для перехода на страницу сессии
  const handleViewSessionDetails = (sessionId) => {
    navigate(`${pathBase}/session/${sessionId}`);
  };

  return (
    <div className={styles.container}>

      {loading && page === 0 ? (
        <p className={`${styles.message} ${themeClass}`}>Загрузка истории сессий...</p>
      ) : error ? (
        <p className={`${styles.message} ${styles.error} ${themeClass}`}>{error}</p>
      ) : sessions.length === 0 ? (
        <Card theme={theme}>
          <p className={`${styles.message} ${themeClass}`}>У вас пока нет истории моек</p>
        </Card>
      ) : (
        <>
          <div className={styles.sessionsList}>
            {sessions.map((session) => (
              <Card key={session.id} theme={theme} className={styles.sessionCard}>
                <div className={styles.sessionHeader}>
                  <StatusBadge status={session.status} theme={theme} />
                  <span className={`${styles.sessionDate} ${themeClass}`}>
                    {formatDate(session.created_at)}
                  </span>
                </div>
                
                <div className={styles.sessionInfo}>
                  <div className={styles.statusInfo}>
                    <div className={`${styles.statusDot} ${styles[session.status]}`}></div>
                    <span className={`${styles.statusText} ${themeClass}`}>
                      {getSessionStatusDescription(session.status, session.cooldown_minutes)}
                    </span>
                  </div>
                  {session.car_number && (
                    <div className={styles.carNumberInfo}>
                      <span className={`${styles.carNumberText} ${themeClass}`}>
                        Номер машины: {session.car_number}
                      </span>
                    </div>
                  )}
                  
                  {/* Информация о платеже */}
                  {session.payment && (
                    <div className={styles.paymentInfo}>
                      <div className={styles.paymentAmount}>
                        <span className={`${styles.paymentText} ${themeClass}`}>
                          💰 Стоимость: {session.main_payment || session.extension_payments ? 
                            formatSessionTotalCost({
                              main_payment: session.main_payment,
                              extension_payments: session.extension_payments || []
                            }) : 
                            session.payment ? `${(session.payment.amount / 100).toFixed(2)} ${session.payment.currency}` : 
                            'Нет данных'}
                        </span>
                      </div>
                      
                      <div className={styles.paymentStatus}>
                        <span className={`${styles.paymentText} ${themeClass}`} style={{
                          color: session.payment.status === 'succeeded' ? '#4CAF50' : 
                                 session.payment.status === 'pending' ? '#FF9800' : 
                                 session.payment.status === 'refunded' ? '#2196F3' : '#F44336',
                          fontWeight: 'bold'
                        }}>
                          {session.payment.status === 'succeeded' ? '✅ Оплачено' :
                           session.payment.status === 'pending' ? '⏳ Ожидает оплаты' :
                           session.payment.status === 'failed' ? '❌ Ошибка оплаты' :
                           session.payment.status === 'refunded' ? '💸 Возвращено' : session.payment.status}
                        </span>
                      </div>
                    </div>
                  )}
                </div>
                
                <Button 
                  theme={theme} 
                  onClick={() => handleViewSessionDetails(session.id)}
                  className={styles.detailsButton}
                >
                  Подробнее
                </Button>
              </Card>
            ))}
          </div>
          
          {hasMore && (
            <div className={styles.loadMoreContainer}>
              <Button 
                theme={theme} 
                onClick={loadMore}
                disabled={loading}
                className={styles.loadMoreButton}
              >
                {loading ? 'Загрузка...' : 'Загрузить еще'}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default SessionHistory;
