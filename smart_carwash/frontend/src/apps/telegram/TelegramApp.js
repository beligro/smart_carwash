import React, { useEffect, useState, Suspense, lazy, useRef } from 'react';
import WebApp from '@twa-dev/sdk';
import { Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import Header from './components/Header';
import WashInfo from './components/WashInfo/WashInfo';
import PaymentPage from './components/PaymentPage';
import ApiService from '../../shared/services/ApiService';
import { getTheme } from '../../shared/styles/theme';
// import { SettingsProvider } from '../../shared/contexts/SettingsContext';

// Ленивая загрузка компонентов
const SessionDetails = lazy(() => import('./components/SessionDetails'));
const SessionHistory = lazy(() => import('./components/SessionHistory'));
const BookingPage = lazy(() => import('./components/BookingPage'));

// Стилизованные компоненты
const AppContainer = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: ${props => props.theme.backgroundColor};
  color: ${props => props.theme.textColor};
  transition: background-color 0.3s ease, color 0.3s ease;
`;

const ContentContainer = styled.div`
  flex: 1;
  padding: 16px;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
`;

/**
 * Компонент приложения Telegram Mini App
 */
const TelegramApp = () => {
  const navigate = useNavigate();
  const location = useLocation();
  // Состояния
  const [washInfo, setWashInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [theme, setTheme] = useState('light');
  const [user, setUser] = useState(null);
  
  // Refs для управления интервалами поллинга
  const queuePollingInterval = useRef(null);
  const sessionPollingInterval = useRef(null);
  const sessionResetTimer = useRef(null); // Таймер для сброса сессии

  // Инициализация приложения
  useEffect(() => {
    const initializeApp = async () => {
      try {
        // Инициализация Telegram Mini App
        WebApp.ready();
        
        // Фиксируем светлую тему (игнорируем тему Telegram)
        setTheme('light');
        
        // Настройка основного цвета (светлая тема)
        document.documentElement.style.setProperty(
          '--tg-theme-button-color', 
          '#2481cc'
        );
        
        // Настройка цвета текста (светлая тема)
        document.documentElement.style.setProperty(
          '--tg-theme-text-color', 
          '#000000'
        );
        
        // Получаем данные пользователя из Telegram
        if (WebApp.initDataUnsafe && WebApp.initDataUnsafe.user) {
          const telegramUser = WebApp.initDataUnsafe.user;
          
          // Получаем пользователя по telegram_id
          await getUserByTelegramId(telegramUser.id);
        }
      } catch (err) {
        alert('Ошибка инициализации Telegram WebApp: ' + err.message);
        setError('Ошибка инициализации приложения');
      }
    };
    
    initializeApp();
  }, []);
  
  // Функция для очистки интервалов поллинга
  const clearPollingIntervals = () => {
    if (queuePollingInterval.current) {
      clearInterval(queuePollingInterval.current);
      queuePollingInterval.current = null;
    }
    if (sessionPollingInterval.current) {
      clearInterval(sessionPollingInterval.current);
      sessionPollingInterval.current = null;
    }
    if (sessionResetTimer.current) {
      clearTimeout(sessionResetTimer.current);
      sessionResetTimer.current = null;
    }
  };

  // Функция для проверки и сброса сессии в терминальном статусе
  const checkAndResetTerminalSession = (session) => {
    // Проверяем, находится ли сессия в терминальном статусе
    if (
      session.status === 'complete' || 
      session.status === 'canceled' ||
      session.status === 'expired' ||
      session.status === 'payment_failed'
    ) {
      // Если мы на главной странице, запускаем таймер сброса
      if (location.pathname === '/telegram' || location.pathname === '/telegram/') {
        // Очищаем старый таймер сброса, если он существует
        if (sessionResetTimer.current) {
          clearTimeout(sessionResetTimer.current);
          sessionResetTimer.current = null;
        }
        
        // Запускаем новый таймер сброса
        sessionResetTimer.current = setTimeout(() => {
          setWashInfo(prevInfo => ({
            ...prevInfo,
            userSession: null,
            payment: null
          }));
          sessionResetTimer.current = null;
        }, 5000);
      }
    }
  };

  // Функция для получения пользователя по telegram_id
  const getUserByTelegramId = async (telegramId) => {
    try {
      const response = await ApiService.getUserByTelegramId(telegramId);
      setUser(response.user);
      return response.user;
    } catch (err) {
      alert('Ошибка получения пользователя: ' + err.message);
      
      // Если это тестовый пользователь, попробуем создать его
      if (telegramId === 12345678) {
        try {
          const createResponse = await ApiService.createUser({
            telegramId: telegramId,
            username: 'test_user',
            firstName: 'Test',
            lastName: 'User'
          });
          setUser(createResponse.user);
          return createResponse.user;
        } catch (createErr) {
          alert('Ошибка создания тестового пользователя: ' + createErr.message);
          setError('Не удалось создать тестового пользователя. Проверьте, что бэкенд запущен и доступен.');
          return null;
        }
      } else {
        setError('Не удалось получить пользователя. Возможно, вы не зарегистрированы в боте. Пожалуйста, нажмите /start в боте.');
        return null;
      }
    }
  };

  // Функция для загрузки статуса очереди и боксов
  const fetchQueueStatus = async (isInitialLoad = false) => {
    try {
      // Устанавливаем loading только при первой загрузке
      if (isInitialLoad) {
        setLoading(true);
      }
      
      const response = await ApiService.getQueueStatus();
      
      if (response) {
        // Обновляем информацию о мойке
        setWashInfo(prevInfo => {
          const newInfo = {
            ...prevInfo,
            allBoxes: response.boxes || [],
            washQueue: response.wash_queue || { queue_size: 0, has_queue: false },
            airDryQueue: response.air_dry_queue || { queue_size: 0, has_queue: false },
            vacuumQueue: response.vacuum_queue || { queue_size: 0, has_queue: false }
          };
          return newInfo;
        });
        
        // Если это первая загрузка, загружаем информацию о сессии пользователя
        if (isInitialLoad && user) {
          try {
            const sessionResponse = await ApiService.getUserSession(user.id);
            
            if (sessionResponse && sessionResponse.session) {
              setWashInfo(prevInfo => {
                const updatedInfo = {
                  ...prevInfo,
                  userSession: sessionResponse.session,
                  payment: sessionResponse.payment // Добавляем информацию о платеже
                };
                return updatedInfo;
              });
              
              // Проверяем, нужно ли запускать поллинг или сброс
              const session = sessionResponse.session;
              if (
                session.status === 'complete' || 
                session.status === 'canceled' ||
                session.status === 'expired' ||
                session.status === 'payment_failed'
              ) {
                // Сессия в терминальном статусе, проверяем нужно ли сбросить
                checkAndResetTerminalSession(session);
              } else {
                // Сессия активна, запускаем поллинг
                startSessionPolling(session.id);
              }
            }
          } catch (err) {
            alert('Ошибка загрузки информации о пользовательской сессии: ' + err.message);
          }
        }
        
        setError(null);
      } else {
        // Если ответ пустой, инициализируем с пустыми данными
        if (isInitialLoad) {
          setWashInfo({
            allBoxes: [],
            washQueue: { queue_size: 0, has_queue: false },
            airDryQueue: { queue_size: 0, has_queue: false },
            vacuumQueue: { queue_size: 0, has_queue: false },
            userSession: null,
            payment: null
          });
        }
      }
    } catch (err) {
      console.error('Ошибка при получении статуса очереди:', err);
      if (isInitialLoad) {
        setError('Ошибка при загрузке данных');
        // Инициализируем washInfo с пустыми данными при ошибке
        setWashInfo({
          allBoxes: [],
          washQueue: { queue_size: 0, has_queue: false },
          airDryQueue: { queue_size: 0, has_queue: false },
          vacuumQueue: { queue_size: 0, has_queue: false },
          userSession: null,
          payment: null
        });
      }
    } finally {
      if (isInitialLoad) {
        setLoading(false);
      }
    }
  };
  
  // Функция для создания сессии с платежом
  const handleCreateSessionWithPayment = async (serviceData) => {
    try {
      setLoading(true);
      setError(null);
      
      // Создаем сессию с платежом
      const requestData = {
        userId: user.id,
        serviceType: serviceData.serviceType,
        withChemistry: serviceData.withChemistry,
        chemistryTimeMinutes: serviceData.chemistryTimeMinutes || 0,
        carNumber: serviceData.carNumber,
        rentalTimeMinutes: serviceData.rentalTimeMinutes
      };
      
      const response = await ApiService.createSessionWithPayment(requestData);

      if (response && response.session && response.payment) {
        // Очищаем старый поллинг сессии перед созданием нового
        if (sessionPollingInterval.current) {
          clearInterval(sessionPollingInterval.current);
          sessionPollingInterval.current = null;
        }
        
        // Обновляем информацию о сессии пользователя
        setWashInfo(prevInfo => ({
          ...prevInfo,
          userSession: response.session,
          payment: response.payment
        }));
        
        // Запускаем поллинг для новой сессии
        startSessionPolling(response.session.id);
        
        // Переходим на страницу оплаты
        navigate('/telegram/payment', {
          state: {
            session: response.session,
            payment: response.payment
          }
        });
      } else {
        setError('Ошибка при создании сессии с платежом');
      }
    } catch (err) {
      alert('Ошибка при создании сессии с платежом: ' + err.message);
      setError('Не удалось создать сессию с платежом. Пожалуйста, попробуйте еще раз.');
    } finally {
      setLoading(false);
    }
  };
  
  // Функция для запуска поллинга статуса очереди и боксов
  const startQueuePolling = () => {
    // Очищаем старый интервал, если он существует
    if (queuePollingInterval.current) {
      clearInterval(queuePollingInterval.current);
    }
    
    // Устанавливаем интервал для поллинга (каждые 10 секунд)
    queuePollingInterval.current = setInterval(() => {
      fetchQueueStatus(false);
    }, 10000);
    
    return queuePollingInterval.current;
  };
  
  // Функция для поллинга статуса сессии
  const startSessionPolling = (sessionId) => {
    // Очищаем старый интервал, если он существует
    if (sessionPollingInterval.current) {
      clearInterval(sessionPollingInterval.current);
    }
    
    // Очищаем старый таймер сброса, если он существует
    if (sessionResetTimer.current) {
      clearTimeout(sessionResetTimer.current);
      sessionResetTimer.current = null;
    }
    
    // Устанавливаем интервал для поллинга (каждые 5 секунд)
    sessionPollingInterval.current = setInterval(async () => {
      try {
        const sessionData = await ApiService.getSessionById(sessionId);
        
        if (sessionData && sessionData.session) {
          // Обновляем данные сессии пользователя
          setWashInfo(prevInfo => {
            return {
              ...prevInfo,
              userSession: sessionData.session,
              payment: sessionData.payment // Добавляем информацию о платеже
            };
          });
          
          // Если сессия завершена, отменена или истекла, останавливаем поллинг
          if (
            sessionData.session.status === 'complete' || 
            sessionData.session.status === 'canceled' ||
            sessionData.session.status === 'expired' ||
            sessionData.session.status === 'payment_failed'
          ) {
            clearInterval(sessionPollingInterval.current);
            sessionPollingInterval.current = null;
            
            // Сброс сессии через 5 секунд только если мы на главной странице
            if (location.pathname === '/telegram' || location.pathname === '/telegram/') {
              sessionResetTimer.current = setTimeout(() => {
                setWashInfo(prevInfo => ({
                  ...prevInfo,
                  userSession: null,
                  payment: null
                }));
                sessionResetTimer.current = null;
              }, 5000);
            }
          }
        }
      } catch (err) {
        console.error('Ошибка при получении статуса сессии:', err);
        // Не показываем alert, просто логируем ошибку
      }
    }, 5000);
    
    return sessionPollingInterval.current;
  };
  
  // Запускаем поллинг при монтировании компонента
  useEffect(() => {
    if (user) {
      const loadData = async () => {
        try {
          // Загружаем статус очереди и боксов при первой загрузке
          await fetchQueueStatus(true);
          
          // Запускаем поллинг статуса очереди (должен работать всегда)
          startQueuePolling();

          // Загружаем информацию о сессии пользователя по user_id и запускаем поллинг по session_id
          const loadUserSessionAndStartPolling = async () => {
            try {
              // Сначала пробуем получить сессию напрямую через getUserSession
              const sessionResponse = await ApiService.getUserSession(user.id);
              
              if (sessionResponse && sessionResponse.session) {
                // Обновляем данные сессии пользователя
                setWashInfo(prevInfo => {
                  return {
                    ...prevInfo,
                    userSession: sessionResponse.session,
                    payment: sessionResponse.payment // Добавляем информацию о платеже
                  };
                });
                
                // Проверяем, нужно ли запускать поллинг или сброс
                const session = sessionResponse.session;
                if (
                  session.status === 'complete' || 
                  session.status === 'canceled' ||
                  session.status === 'expired' ||
                  session.status === 'payment_failed'
                ) {
                  // Сессия в терминальном статусе, проверяем нужно ли сбросить
                  checkAndResetTerminalSession(session);
                } else {
                  // Сессия активна, запускаем поллинг
                  startSessionPolling(session.id);
                }
              }
            } catch (err) {
              console.error('Ошибка загрузки информации о пользовательской сессии:', err);
            }
          };
          
          loadUserSessionAndStartPolling();
        } catch (err) {
          console.error('Error in useEffect for polling:', err);
          setError('Ошибка при загрузке данных');
          setLoading(false);
        }
      };
      
      loadData();
    }
  }, [user]);

  // Очистка интервалов при размонтировании компонента
  useEffect(() => {
    return () => {
      clearPollingIntervals();
    };
  }, []);

  // Проверяем сессию при изменении маршрута на главную страницу
  useEffect(() => {
    if (location.pathname === '/telegram' || location.pathname === '/telegram/') {
      // Если у нас есть сессия в терминальном статусе, запускаем таймер сброса
      if (washInfo && washInfo.userSession) {
        checkAndResetTerminalSession(washInfo.userSession);
      }
    }
  }, [location.pathname, washInfo?.userSession]);

  // Функция для перехода на страницу истории сессий
  const handleViewHistory = () => {
    navigate('/telegram/history');
  };

  // Функция для возврата на главную страницу
  const handleBackToHome = () => {
    navigate('/telegram');
  };

  // Обработчики для страницы оплаты
  const handlePaymentComplete = (updatedSession) => {
    // Очищаем старый поллинг сессии перед обновлением
    if (sessionPollingInterval.current) {
      clearInterval(sessionPollingInterval.current);
      sessionPollingInterval.current = null;
    }
    
    // Платеж успешен, обновляем информацию о сессии
    setWashInfo(prevInfo => ({
      ...prevInfo,
      userSession: updatedSession,
      payment: updatedSession.payment
    }));
    
    // Запускаем поллинг для обновленной сессии
    startSessionPolling(updatedSession.id);
    
    // Всегда возвращаемся на главную страницу mini app после успешной оплаты
    navigate('/telegram');
  };

  const handlePaymentFailed = (updatedSession) => {
    // Очищаем старый поллинг сессии перед обновлением
    if (sessionPollingInterval.current) {
      clearInterval(sessionPollingInterval.current);
      sessionPollingInterval.current = null;
    }
    
    // Платеж неудачен, обновляем информацию о сессии
    setWashInfo(prevInfo => ({
      ...prevInfo,
      userSession: updatedSession,
      payment: updatedSession.payment
    }));
    
    // Запускаем поллинг для обновленной сессии
    startSessionPolling(updatedSession.id);
    
    // Возвращаемся на главную страницу
    navigate('/telegram');
  };

  const handlePaymentBack = () => {
    // Всегда возвращаемся на главную страницу mini app
    navigate('/telegram');
  };

    // Обработчик отмены сессии
  const handleCancelSession = async (sessionId, userId) => {
    try {
      // Отменяем сессию через API
      const response = await ApiService.cancelSession(sessionId, userId);
      
      // Немедленно обновляем данные сессии для мгновенного отображения изменений
      try {
        const updatedSessionData = await ApiService.getSessionById(sessionId);
        if (updatedSessionData && updatedSessionData.session) {
          // Обновляем информацию о сессии с актуальными данными
          setWashInfo(prevInfo => ({
            ...prevInfo,
            userSession: updatedSessionData.session,
            payment: updatedSessionData.payment
          }));
        }
      } catch (refreshError) {
        console.error('Ошибка при обновлении данных сессии:', refreshError);
        // Fallback к данным из ответа API
        setWashInfo(prevInfo => ({
          ...prevInfo,
          userSession: response.session,
          payment: response.payment
        }));
      }
      
      // Очищаем поллинг сессии, так как она отменена
      if (sessionPollingInterval.current) {
        clearInterval(sessionPollingInterval.current);
        sessionPollingInterval.current = null;
      }
      
      // Очищаем старый таймер сброса, если он существует
      if (sessionResetTimer.current) {
        clearTimeout(sessionResetTimer.current);
        sessionResetTimer.current = null;
      }
      
      // Сбрасываем сессию через 5 секунд на главной странице
      if (location.pathname === '/telegram' || location.pathname === '/telegram/') {
        sessionResetTimer.current = setTimeout(() => {
          setWashInfo(prevInfo => ({
            ...prevInfo,
            userSession: null,
            payment: null
          }));
          sessionResetTimer.current = null;
        }, 5000);
      }
      
      // Запускаем поллинг очереди для обновления статуса
      startQueuePolling();
      
      console.log('Сессия успешно отменена:', response);
    } catch (error) {
      console.error('Ошибка при отмене сессии:', error);
      throw error;
    }
  };

  // Обработчик завершения сессии
  const handleCompleteSession = (updatedSession, updatedPayment) => {
    // Обновляем информацию о сессии с актуальными данными
    setWashInfo(prevInfo => ({
      ...prevInfo,
      userSession: updatedSession,
      payment: updatedPayment
    }));
  };

  const themeObject = getTheme(theme);

  // Обработка ошибок рендеринга
  if (!themeObject) {
    alert('Ошибка: не удалось получить тему');
    return <div>Ошибка загрузки приложения</div>;
  }

  // Определяем, нужно ли показывать стрелку назад
  const showBackButton = location.pathname !== '/telegram/' && location.pathname !== '/telegram';

  return (
    <AppContainer theme={themeObject}>
      <Header theme={theme} onBack={showBackButton ? handleBackToHome : undefined} />
        <ContentContainer>
          <Routes>
            <Route 
              path="/" 
              element={
                <>
                  {loading ? (
                    <p>Загрузка информации о мойке...</p>
                  ) : error ? (
                    <p style={{ color: 'red' }}>{error}</p>
                  ) : washInfo ? (
                    <WashInfo 
                      washInfo={washInfo} 
                      theme={theme} 
                      onCreateSession={handleCreateSessionWithPayment}
                      onViewHistory={handleViewHistory}
                      onCancelSession={handleCancelSession}
                      onCompleteSession={handleCompleteSession}
                      user={user}
                    />
                  ) : (
                    <p>Нет данных для отображения</p>
                  )}
                </>
              } 
            />
            <Route 
              path="/session/:sessionId" 
              element={
                <Suspense fallback={<div>Загрузка информации о сессии...</div>}>
                  <SessionDetails 
                    theme={theme} 
                    user={user}
                  />
                </Suspense>
              } 
            />
            <Route 
              path="/history" 
              element={
                <Suspense fallback={<div>Загрузка истории сессий...</div>}>
                  <SessionHistory 
                    theme={theme} 
                    user={user}
                  />
                </Suspense>
              } 
            />
            <Route 
              path="/payment"
              element={
                <Suspense fallback={<div>Загрузка страницы оплаты...</div>}>
                  <PaymentPage 
                    session={location?.state?.session}
                    payment={location?.state?.payment}
                    onPaymentComplete={handlePaymentComplete}
                    onPaymentFailed={handlePaymentFailed}
                    onBack={handlePaymentBack}
                    theme={theme}
                    paymentType={location?.state?.paymentType || 'main'}
                  />
                </Suspense>
              } 
            />
            <Route 
              path="/booking"
              element={
                <Suspense fallback={<div>Загрузка страницы записи...</div>}>
                  <BookingPage 
                    theme={theme}
                    user={user}
                    onCreateSession={handleCreateSessionWithPayment}
                  />
                </Suspense>
              } 
            />
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </ContentContainer>
      </AppContainer>
  );
};

export default TelegramApp;
