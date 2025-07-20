import React, { useEffect, useState, Suspense, lazy } from 'react';
import WebApp from '@twa-dev/sdk';
import { Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import Header from './components/Header';
import WelcomeMessage from './components/WelcomeMessage/WelcomeMessage';
import WashInfo from './components/WashInfo/WashInfo';
import PaymentPage from './components/PaymentPage';
import ApiService from '../../shared/services/ApiService';
import { getTheme } from '../../shared/styles/theme';

// Ленивая загрузка компонентов
const SessionDetails = lazy(() => import('./components/SessionDetails'));
const SessionHistory = lazy(() => import('./components/SessionHistory'));

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

  // Инициализация приложения
  useEffect(() => {
    const initializeApp = async () => {
      try {
        // Инициализация Telegram Mini App
        WebApp.ready();
        
        // Установка темы в соответствии с темой Telegram
        const colorScheme = WebApp.colorScheme || 'light';
        setTheme(colorScheme);
        
        // Настройка основного цвета
        document.documentElement.style.setProperty(
          '--tg-theme-button-color', 
          (WebApp.themeParams && WebApp.themeParams.button_color) || '#2481cc'
        );
        
        // Настройка цвета текста
        document.documentElement.style.setProperty(
          '--tg-theme-text-color', 
          (WebApp.themeParams && WebApp.themeParams.text_color) || '#000000'
        );
        
        // Получаем данные пользователя из Telegram
        if (WebApp.initDataUnsafe && WebApp.initDataUnsafe.user) {
          const telegramUser = WebApp.initDataUnsafe.user;
          
          // Получаем пользователя по telegram_id
          getUserByTelegramId(telegramUser.id);
        }
      } catch (err) {
        console.error('Ошибка инициализации Telegram WebApp:', err);
      }
    };
    
    initializeApp();
  }, []);
  
  // Функция для получения пользователя по telegram_id
  const getUserByTelegramId = async (telegramId) => {
    try {
      const response = await ApiService.getUserByTelegramId(telegramId);
      setUser(response.user);
    } catch (err) {
      console.error('Ошибка получения пользователя:', err);
      
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
        } catch (createErr) {
          console.error('Ошибка создания тестового пользователя:', createErr);
          setError('Не удалось создать тестового пользователя. Проверьте, что бэкенд запущен и доступен.');
        }
      } else {
        setError('Не удалось получить пользователя. Возможно, вы не зарегистрированы в боте. Пожалуйста, нажмите /start в боте.');
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
      
      const data = await ApiService.getQueueStatus();
      
      // Обновляем данные, сохраняя структуру объекта
      setWashInfo(prevInfo => {
        // Если это первая загрузка или предыдущих данных нет
        if (!prevInfo) {
          const newInfo = {
            allBoxes: data.all_boxes || [],
            washQueue: data.wash_queue || { queue_size: 0, has_queue: false },
            airDryQueue: data.air_dry_queue || { queue_size: 0, has_queue: false },
            vacuumQueue: data.vacuum_queue || { queue_size: 0, has_queue: false },
            totalQueueSize: data.total_queue_size || 0,
            hasAnyQueue: data.has_any_queue || false,
            userSession: data.user_session || null
          };
          return newInfo;
        }
        
        // Обновляем только изменившиеся данные
        const updatedInfo = {
          ...prevInfo,
          allBoxes: data.all_boxes || prevInfo.allBoxes,
          washQueue: data.wash_queue || prevInfo.washQueue,
          airDryQueue: data.air_dry_queue || prevInfo.airDryQueue,
          vacuumQueue: data.vacuum_queue || prevInfo.vacuumQueue,
          totalQueueSize: data.total_queue_size || prevInfo.totalQueueSize,
          hasAnyQueue: data.has_any_queue || prevInfo.hasAnyQueue,
          userSession: data.user_session || prevInfo.userSession
        };
        return updatedInfo;
      });
      
      setError(null);
    } catch (err) {
      console.error('Ошибка загрузки информации о мойке:', err);
      setError('Не удалось загрузить информацию о мойке');
      
      // Создаем пустой объект с необходимой структурой только если нет предыдущих данных
      if (!washInfo) {
        setWashInfo({
          allBoxes: [],
          washQueue: { serviceType: 'wash', boxes: [], queueSize: 0, hasQueue: false },
          airDryQueue: { serviceType: 'air_dry', boxes: [], queueSize: 0, hasQueue: false },
          vacuumQueue: { serviceType: 'vacuum', boxes: [], queueSize: 0, hasQueue: false },
          totalQueueSize: 0,
          hasAnyQueue: false,
          userSession: null
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
      if (!user) {
        setError('Пользователь не авторизован');
        return;
      }
      
      setLoading(true);
      
      // Проверяем, что все необходимые данные присутствуют
      if (!serviceData.serviceType || !serviceData.carNumber) {
        setError('Не все данные заполнены');
        return;
      }
      
      // Создаем сессию с платежом
      const response = await ApiService.createSessionWithPayment({ 
        userId: user.id,
        serviceType: serviceData.serviceType,
        withChemistry: serviceData.withChemistry || false,
        rentalTimeMinutes: serviceData.rentalTimeMinutes || 5,
        carNumber: serviceData.carNumber
      });

      if (response && response.session && response.payment) {
        // Переходим на страницу оплаты
        navigate('/telegram/payment', { 
          state: { 
            session: response.session, 
            payment: response.payment 
          } 
        });
      } else {
        alert('Invalid response structure: ' + JSON.stringify(response));
        setError('Ошибка создания сессии с платежом: неверная структура ответа');
      }
    } catch (err) {
      alert('Ошибка создания сессии с платежом: ' + err.message);
      alert('Error details: ' + JSON.stringify(err.response?.data || err.message));
      setError(`Не удалось создать сессию с платежом: ${err.response?.data?.error || err.message}`);
    } finally {
      setLoading(false);
    }
  };

  // Функция для создания сессии (старый метод для совместимости)
  const handleCreateSession = async (serviceData) => {
    try {
      if (!user) {
        setError('Пользователь не авторизован');
        return;
      }
      
      setLoading(true);
      
      // Проверяем, что все необходимые данные присутствуют
      if (!serviceData.serviceType || !serviceData.carNumber) {
        setError('Не все данные заполнены');
        return;
      }
      
      // Добавляем данные о типе услуги, химии и номере машины в запрос
      const response = await ApiService.createSession({ 
        user_id: user.id,
        service_type: serviceData.serviceType,
        with_chemistry: serviceData.withChemistry || false,
        rental_time_minutes: serviceData.rentalTimeMinutes || 5,
        car_number: serviceData.carNumber
      });
      
      if (response.session) {
        // Переходим на страницу сессии
        navigate(`/telegram/session/${response.session.id}`);
        
        // Запускаем поллинг для обновления статуса сессии
        startSessionPolling(response.session.id);
      } else {
        setError('Не удалось создать сессию');
      }
    } catch (err) {
      console.error('Ошибка создания сессии:', err);
      setError('Не удалось создать сессию. Попробуйте еще раз.');
    } finally {
      setLoading(false);
    }
  };
  
  // Функция для запуска поллинга статуса очереди и боксов
  const startQueuePolling = () => {
    // Устанавливаем интервал для поллинга (каждые 10 секунд)
    const queuePollingInterval = setInterval(() => {
      fetchQueueStatus(false);
    }, 10000);
    
    // Сохраняем интервал в ref, чтобы можно было его очистить при размонтировании компонента
    // Для простоты в этом примере мы не очищаем интервал
    return queuePollingInterval;
  };
  
  // Функция для запуска поллинга статуса сессии
  const startSessionPolling = (sessionId) => {
    // Устанавливаем интервал для поллинга (каждые 5 секунд)
    const sessionPollingInterval = setInterval(async () => {
      try {
        const sessionData = await ApiService.getSessionById(sessionId);
        
        if (sessionData && sessionData.session) {
          // Обновляем данные сессии пользователя
          setWashInfo(prevInfo => {
            return {
              ...prevInfo,
              userSession: sessionData.session
            };
          });
          
          // Если сессия завершена, отменена или истекла, останавливаем поллинг
          if (
            sessionData.session.status === 'complete' || 
            sessionData.session.status === 'canceled' ||
            sessionData.session.status === 'expired'
          ) {
            clearInterval(sessionPollingInterval);
            
            // Через 5 секунд обновляем блок с сессией и предлагаем начать новую
            setTimeout(() => {
              setWashInfo(prevInfo => ({
                ...prevInfo,
                userSession: null
              }));
            }, 5000);
          }
        }
      } catch (err) {
        console.error('Ошибка при получении статуса сессии:', err);
      }
    }, 5000);
    
    // Сохраняем интервал в ref, чтобы можно было его очистить при размонтировании компонента
    // Для простоты в этом примере мы не очищаем интервал
    return sessionPollingInterval;
  };
  
  // Запускаем поллинг при монтировании компонента
  useEffect(() => {
    if (user) {
      try {
        // Загружаем статус очереди и боксов при первой загрузке
        fetchQueueStatus(true);
        
        // Запускаем поллинг статуса очереди (должен работать всегда)
        const queueInterval = startQueuePolling();

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
                  userSession: sessionResponse.session
                };
              });
              
              // Если у пользователя есть активная сессия, запускаем поллинг для неё по ID сессии
              if (['created', 'in_queue', 'payment_failed', 'assigned', 'active'].includes(sessionResponse.session.status)) {
                startSessionPolling(sessionResponse.session.id);
              }
            } else {
              console.log('No user session found');
            }
          } catch (err) {
            console.error('Ошибка загрузки информации о пользовательской сессии:', err);
          }
        };
        
        loadUserSessionAndStartPolling();
        
        // Очищаем интервал при размонтировании
        return () => {
          clearInterval(queueInterval);
        };
      } catch (err) {
        console.error('Error in useEffect for polling:', err);
        setError('Ошибка при загрузке данных');
        setLoading(false);
      }
    } else {
      console.log('No user yet, skipping data loading');
    }
  }, [user]);

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
    // Платеж успешен, переходим к деталям сессии
    navigate('/telegram/session/' + updatedSession.id, { state: { session: updatedSession } });
  };

  const handlePaymentFailed = (updatedSession) => {
    // Платеж неудачен, остаемся на странице оплаты
    // Обновляем данные сессии
    setWashInfo(prevInfo => ({
      ...prevInfo,
      userSession: updatedSession
    }));
  };

  const handlePaymentBack = () => {
    // Возвращаемся на главную страницу
    navigate('/telegram');
  };

  const themeObject = getTheme(theme);

  // Обработка ошибок рендеринга
  if (!themeObject) {
    console.error('Ошибка: не удалось получить тему');
    return <div>Ошибка загрузки приложения</div>;
  }

  return (
    <AppContainer theme={themeObject}>
      <Header theme={theme} />
        <ContentContainer>
          <Routes>
            <Route 
              path="/" 
              element={
                <>
                  <WelcomeMessage theme={theme} />
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
                    onBack={handleBackToHome}
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
