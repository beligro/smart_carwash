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
  
  // Получаем данные из washInfo (поддерживаем оба формата)
  const allBoxes = washInfo?.allBoxes || washInfo?.all_boxes || [];
  const washQueue = washInfo?.washQueue || washInfo?.wash_queue || { queue_size: 0, has_queue: false };
  const airDryQueue = washInfo?.airDryQueue || washInfo?.air_dry_queue || { queue_size: 0, has_queue: false };
  const vacuumQueue = washInfo?.vacuumQueue || washInfo?.vacuum_queue || { queue_size: 0, has_queue: false };
  const totalQueueSize = washInfo?.totalQueueSize || washInfo?.total_queue_size || 0;
  const hasAnyQueue = washInfo?.hasAnyQueue || washInfo?.has_any_queue || false;
  
  // Используем userSession из washInfo
  const userSession = washInfo?.userSession || washInfo?.user_session;
  
  // Используем хук для таймера
  const { timeLeft } = useTimer(userSession);
  
  // Функция для перехода на страницу сессии
  const handleViewSessionDetails = () => {
    try {
      if (userSession && userSession.id) {
        navigate(`/telegram/session/${userSession.id}`);
      }
    } catch (error) {
      console.error('Ошибка при переходе к деталям сессии:', error);
    }
  };

  // Обработчик нажатия на кнопку "Записаться на мойку"
  const handleCreateSessionClick = () => {
    try {
      setShowServiceSelector(true);
    } catch (error) {
      console.error('Ошибка при открытии выбора услуг:', error);
    }
  };

  // Обработчик выбора услуги
  const handleServiceSelect = (serviceData) => {
    try {
      setShowServiceSelector(false);
      onCreateSession(serviceData);
    } catch (error) {
      console.error('Ошибка при выборе услуги:', error);
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
              {/* Отладочная информация */}
              <p className={`${styles.sessionInfo} ${themeClass}`} style={{ fontSize: '12px', color: '#666' }}>
                Отладка: status={userSession.status}, serviceType={userSession.serviceType || userSession.service_type}, 
                createdAt={userSession.createdAt || userSession.created_at}
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
              
              {/* Отображаем таймер для активной сессии */}
              {userSession.status === 'active' && timeLeft !== null && (
                <>
                  <p className={`${styles.sessionInfo} ${themeClass}`} style={{ marginTop: '12px', fontWeight: 'bold' }}>
                    Оставшееся время мойки:
                  </p>
                  <Timer seconds={timeLeft} theme={theme} />
                </>
              )}
              
              {/* Отладочная информация для таймера */}
              <p className={`${styles.sessionInfo} ${themeClass}`} style={{ fontSize: '12px', color: '#666' }}>
                Отладка таймера: status={userSession.status}, timeLeft={timeLeft}, 
                status_updated_at={userSession.status_updated_at || 'null'}, 
                updated_at={userSession.updated_at || 'null'},
                createdAt={userSession.created_at || 'null'}
              </p>
              
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

      {/* Информация о боксах */}
      <section className={styles.section}>
        <h2 className={`${styles.title} ${themeClass}`}>Боксы автомойки</h2>
        {allBoxes.length > 0 ? (
          <>
            <h3 className={`${styles.subtitle} ${themeClass}`}>Мойка</h3>
            <div className={styles.boxesGrid}>
              {allBoxes
                .filter(box => (box.serviceType || box.service_type) === 'wash')
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
                .filter(box => (box.serviceType || box.service_type) === 'air_dry')
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
                .filter(box => (box.serviceType || box.service_type) === 'vacuum')
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
