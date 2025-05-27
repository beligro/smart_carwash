import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './SessionHistory.module.css';
import { Card, Button, StatusBadge } from '../UI';
import { formatDate } from '../../utils/formatters';
import { getSessionStatusDescription } from '../../utils/statusHelpers';
import ApiService from '../../services/ApiService';

/**
 * Компонент SessionHistory - отображает историю сессий пользователя
 * @param {Object} props - Свойства компонента
 * @param {Object} props.user - Информация о пользователе
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Function} props.onBack - Функция для возврата на главную страницу
 */
const SessionHistory = ({ user, theme = 'light', onBack }) => {
  const navigate = useNavigate();
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
        const response = await ApiService.getUserSessionHistory(user.id, limit, offset);
        
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
        console.error('Ошибка при загрузке истории сессий:', err);
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
    navigate(`/session/${sessionId}`);
  };

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2 className={`${styles.title} ${themeClass}`}>История моек</h2>
        <Button theme={theme} onClick={onBack} className={styles.backButton}>
          Вернуться на главную
        </Button>
      </div>

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
                      {getSessionStatusDescription(session.status)}
                    </span>
                  </div>
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
