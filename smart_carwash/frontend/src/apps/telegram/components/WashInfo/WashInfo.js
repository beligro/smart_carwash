import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './WashInfo.module.css';
import { Card, Button, StatusBadge, Timer } from '../../../../shared/components/UI';
import ServiceSelector from '../ServiceSelector';
import { formatDate } from '../../../../shared/utils/formatters';
import { getSessionStatusDescription, getServiceTypeDescription } from '../../../../shared/utils/statusHelpers';
import useTimer from '../../../../shared/hooks/useTimer';

/**
 * Компонент WashInfo - отображает информацию о мойке
 * @param {Object} props - Свойства компонента
 * @param {Object} props.washInfo - Информация о мойке
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Function} props.onCreateSession - Функция для создания сессии
 * @param {Function} props.onViewHistory - Функция для просмотра истории сессий
 * @param {Object} props.user - Данные пользователя
 */
const WashInfo = ({ washInfo, theme = 'light', onCreateSession, onViewHistory, user }) => {
  const navigate = useNavigate();
  const [showServiceSelector, setShowServiceSelector] = useState(false);
  
  // Получаем данные из washInfo (теперь в camelCase)
  const allBoxes = washInfo?.allBoxes || [];
  const washQueue = washInfo?.washQueue || { queueSize: 0, hasQueue: false };
  const airDryQueue = washInfo?.airDryQueue || { queueSize: 0, hasQueue: false };
  const vacuumQueue = washInfo?.vacuumQueue || { queueSize: 0, hasQueue: false };
  const totalQueueSize = washInfo?.totalQueueSize || 0;
  const hasAnyQueue = washInfo?.hasAnyQueue || false;
  
  // Используем userSession из washInfo
  const userSession = washInfo?.userSession;
  
  // Используем хук для таймера
  const { timeLeft } = useTimer(userSession);
  
  // Функция для перехода на страницу сессии
  const handleViewSessionDetails = () => {
    if (userSession && userSession.id) {
      navigate(`/telegram/session/${userSession.id}`);
    }
  };

  // Обработчик нажатия на кнопку "Записаться на мойку"
  const handleCreateSessionClick = () => {
    setShowServiceSelector(true);
  };

  // Обработчик выбора услуги
  const handleServiceSelect = (serviceData) => {
    setShowServiceSelector(false);
    onCreateSession(serviceData);
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
                status={washQueue.hasQueue ? 'busy' : 'free'} 
                theme={theme}
                text={washQueue.hasQueue ? `В очереди: ${washQueue.queueSize}` : 'Нет очереди'}
              />
            </div>
            <div className={styles.queueTypeItem}>
              <h4 className={`${styles.queueTypeTitle} ${themeClass}`}>Обдув</h4>
              <StatusBadge 
                status={airDryQueue.hasQueue ? 'busy' : 'free'} 
                theme={theme}
                text={airDryQueue.hasQueue ? `В очереди: ${airDryQueue.queueSize}` : 'Нет очереди'}
              />
            </div>
            <div className={styles.queueTypeItem}>
              <h4 className={`${styles.queueTypeTitle} ${themeClass}`}>Пылесос</h4>
              <StatusBadge 
                status={vacuumQueue.hasQueue ? 'busy' : 'free'} 
                theme={theme}
                text={vacuumQueue.hasQueue ? `В очереди: ${vacuumQueue.queueSize}` : 'Нет очереди'}
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
                Создана: {formatDate(userSession.createdAt)}
              </p>
              <p className={`${styles.sessionInfo} ${themeClass}`}>
                Услуга: {getServiceTypeDescription(userSession.serviceType)}
                {userSession.withChemistry && ' (с химией)'}
              </p>
              {(userSession.boxId || userSession.boxNumber) && (
                <p className={`${styles.sessionInfo} ${themeClass}`}>
                  Назначен бокс: #{
                    // Используем номер бокса из сессии, если он есть
                    userSession.boxNumber || 
                    // Иначе находим номер бокса по его ID
                    allBoxes.find(box => box.id === userSession.boxId)?.number || 
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
                onClick={handleCreateSessionClick}
                disabled={hasAnyQueue && totalQueueSize > 5}
              >
                Записаться на мойку
              </Button>
              {hasAnyQueue && totalQueueSize > 5 && (
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
        {allBoxes.length > 0 ? (
          <>
            <h3 className={`${styles.subtitle} ${themeClass}`}>Мойка</h3>
            <div className={styles.boxesGrid}>
              {allBoxes
                .filter(box => box.serviceType === 'wash')
                .map((box) => (
                  <Card key={box.id} theme={theme} className={styles.boxCard}>
                    <h3 className={`${styles.boxNumber} ${themeClass}`}>Бокс #{box.number}</h3>
                    <StatusBadge status={box.status} theme={theme} />
                  </Card>
                ))}
            </div>
            
            <h3 className={`${styles.subtitle} ${themeClass}`}>Обдув</h3>
            <div className={styles.boxesGrid}>
              {allBoxes
                .filter(box => box.serviceType === 'air_dry')
                .map((box) => (
                  <Card key={box.id} theme={theme} className={styles.boxCard}>
                    <h3 className={`${styles.boxNumber} ${themeClass}`}>Бокс #{box.number}</h3>
                    <StatusBadge status={box.status} theme={theme} />
                  </Card>
                ))}
            </div>
            
            <h3 className={`${styles.subtitle} ${themeClass}`}>Пылесос</h3>
            <div className={styles.boxesGrid}>
              {allBoxes
                .filter(box => box.serviceType === 'vacuum')
                .map((box) => (
                  <Card key={box.id} theme={theme} className={styles.boxCard}>
                    <h3 className={`${styles.boxNumber} ${themeClass}`}>Бокс #{box.number}</h3>
                    <StatusBadge status={box.status} theme={theme} />
                  </Card>
                ))}
            </div>
          </>
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
