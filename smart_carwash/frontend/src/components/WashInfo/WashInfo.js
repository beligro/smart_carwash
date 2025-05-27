import React from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './WashInfo.module.css';
import { Card, Button, StatusBadge, Timer } from '../UI';
import { formatDate } from '../../utils/formatters';
import { getSessionStatusDescription } from '../../utils/statusHelpers';
import useTimer from '../../hooks/useTimer';

/**
 * Компонент WashInfo - отображает информацию о мойке
 * @param {Object} props - Свойства компонента
 * @param {Object} props.washInfo - Информация о мойке
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Function} props.onCreateSession - Функция для создания сессии
 * @param {Function} props.onViewHistory - Функция для просмотра истории сессий
 */
const WashInfo = ({ washInfo, theme = 'light', onCreateSession, onViewHistory }) => {
  const navigate = useNavigate();
  const boxes = washInfo?.boxes || [];
  const queueSize = washInfo?.queueSize || 0;
  const hasQueue = washInfo?.hasQueue || false;
  
  // Используем userSession из washInfo
  const userSession = washInfo?.userSession || washInfo?.user_session;
  
  // Используем хук для таймера
  const { timeLeft } = useTimer(userSession);
  
  // Функция для перехода на страницу сессии
  const handleViewSessionDetails = () => {
    if (userSession && userSession.id) {
      navigate(`/session/${userSession.id}`);
    }
  };

  const themeClass = theme === 'dark' ? styles.dark : styles.light;

  return (
    <div className={styles.container}>
      {/* Информация об очереди */}
      <section className={styles.section}>
        <h2 className={`${styles.title} ${themeClass}`}>Статус автомойки</h2>
        <Card theme={theme}>
          <div className={styles.infoRow}>
            <div className={styles.infoLabel}>Статус очереди:</div>
            <div className={styles.infoValue}>
              <StatusBadge 
                status={hasQueue ? 'busy' : 'free'} 
                theme={theme}
                text={hasQueue ? 'Есть очередь' : 'Нет очереди'}
              />
            </div>
          </div>
          {hasQueue && (
            <p className={`${styles.sessionInfo} ${themeClass}`}>
              Количество человек в очереди: {queueSize}
            </p>
          )}
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
                Создана: {formatDate(userSession.created_at)}
              </p>
              {userSession.box_id && (
                <p className={`${styles.sessionInfo} ${themeClass}`}>
                  Назначен бокс: #{
                    // Находим номер бокса по его ID
                    boxes.find(box => box.id === userSession.box_id)?.number || 
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
              <Button 
                theme={theme} 
                onClick={handleViewSessionDetails}
              >
                Подробнее о сессии
              </Button>
            </>
          ) : (
            <>
              <p className={`${styles.sessionInfo} ${themeClass}`}>
                У вас нет активной записи на мойку
              </p>
              <Button 
                theme={theme} 
                onClick={onCreateSession}
                disabled={hasQueue && queueSize > 5}
              >
                Записаться на мойку
              </Button>
              {hasQueue && queueSize > 5 && (
                <p className={`${styles.sessionInfo} ${themeClass}`} style={{ marginTop: '8px', color: '#C62828' }}>
                  Очередь слишком большая, попробуйте позже
                </p>
              )}
            </>
          )}
        </Card>
      </section>

      {/* Информация о боксах */}
      <section className={styles.section}>
        <h2 className={`${styles.title} ${themeClass}`}>Боксы автомойки</h2>
        {boxes.length > 0 ? (
          <div className={styles.boxesGrid}>
            {boxes.map((box) => (
              <Card key={box.id} theme={theme} className={styles.boxCard}>
                <h3 className={`${styles.boxNumber} ${themeClass}`}>Бокс #{box.number}</h3>
                <StatusBadge status={box.status} theme={theme} />
              </Card>
            ))}
          </div>
        ) : (
          <Card theme={theme}>
            <p className={`${styles.sessionInfo} ${themeClass}`} style={{ textAlign: 'center' }}>
              Информация о боксах отсутствует
            </p>
          </Card>
        )}
      </section>
    </div>
  );
};

export default WashInfo;
