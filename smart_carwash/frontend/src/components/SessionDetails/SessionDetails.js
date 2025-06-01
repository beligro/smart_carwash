import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styles from './SessionDetails.module.css';
import { Card, Button, StatusBadge, Timer } from '../UI';
import { formatDate } from '../../utils/formatters';
import { getServiceTypeDescription } from '../../utils/statusHelpers';
import ApiService from '../../services/ApiService';
import useTimer from '../../hooks/useTimer';

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
  const [box, setBox] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [showExtendModal, setShowExtendModal] = useState(false);
  const [availableRentalTimes, setAvailableRentalTimes] = useState([]);
  const [selectedExtensionTime, setSelectedExtensionTime] = useState(null);
  const [loadingRentalTimes, setLoadingRentalTimes] = useState(false);
  
  // Используем хук для таймера
  const { timeLeft } = useTimer(session);
  
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
      console.error('Ошибка при загрузке доступного времени аренды:', err);
      setError('Не удалось загрузить доступное время аренды');
    } finally {
      setLoadingRentalTimes(false);
    }
  };

  // Функция для продления сессии
  const handleExtendSession = async () => {
    if (!selectedExtensionTime) {
      setError('Выберите время продления');
      return;
    }

    try {
      setActionLoading(true);
      setError(null);
      
      // Вызываем API для продления сессии
      const response = await ApiService.extendSession(sessionId, selectedExtensionTime);
      
      if (response && response.session) {
        setSession(response.session);
        setShowExtendModal(false); // Закрываем модальное окно
      }
    } catch (err) {
      console.error('Ошибка при продлении сессии:', err);
      setError('Не удалось продлить сессию. Пожалуйста, попробуйте еще раз.');
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
      console.error('Ошибка при завершении сессии:', err);
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
      console.error('Ошибка при загрузке данных о сессии:', err);
      setError('Не удалось загрузить данные о сессии');
      return null;
    } finally {
      setLoading(false);
    }
  };
  
  // Начальная загрузка данных о сессии
  useEffect(() => {
    if (sessionId) {
      fetchSessionDetails();
    }
  }, [sessionId]);
  
  // Настройка поллинга для обновления данных о сессии
  // Примечание: Мы не запускаем поллинг здесь, так как он уже запущен в App.js
  // Вместо этого просто обновляем данные при первой загрузке и при изменении статуса сессии
  useEffect(() => {
    // Обновляем данные каждый раз, когда меняется sessionId
    if (sessionId) {
      fetchSessionDetails();
    }
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
      console.error('Ошибка при запуске сессии:', err);
      setError('Не удалось запустить сессию. Пожалуйста, попробуйте еще раз.');
    } finally {
      setActionLoading(false);
    }
  };
  
  // Функция для возврата на главную страницу
  const handleBack = () => {
    navigate('/');
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
