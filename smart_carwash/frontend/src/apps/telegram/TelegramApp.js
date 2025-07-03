import React, { useEffect, useState, Suspense, lazy } from 'react';
import WebApp from '@twa-dev/sdk';
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import Header from './components/Header';
import WelcomeMessage from './components/WelcomeMessage/WelcomeMessage';
import WashInfo from './components/WashInfo/WashInfo';
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
        console.log('Starting TelegramApp initialization...');
        
        // Проверяем доступность API
        console.log('Checking API health...');
        const apiHealth = await ApiService.checkApiHealth();
        if (!apiHealth) {
          setError('API недоступен. Проверьте, что бэкенд запущен на http://localhost:8080');
          return;
        }
        console.log('API is healthy');
        
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
        
        console.log('Telegram WebApp инициализирован успешно');
        
        // Получаем данные пользователя из Telegram
        if (WebApp.initDataUnsafe && WebApp.initDataUnsafe.user) {
          const telegramUser = WebApp.initDataUnsafe.user;
          console.log('Telegram user found:', telegramUser);
          
          // Получаем пользователя по telegram_id
          getUserByTelegramId(telegramUser.id);
        } else {
          console.log('No Telegram user, using test user');
          // Для разработки используем тестового пользователя
          getUserByTelegramId(12345678);
        }
      } catch (err) {
        console.error('Ошибка инициализации Telegram WebApp:', err);
        console.log('Продолжаем работу в режиме разработки');
        
        // Для разработки используем тестового пользователя
        getUserByTelegramId(12345678);
      }
    };
    
    initializeApp();
  }, []);
  
  // Функция для получения пользователя по telegram_id
  const getUserByTelegramId = async (telegramId) => {
    try {
      console.log('Getting user by telegram ID:', telegramId);
      const response = await ApiService.getUserByTelegramId(telegramId);
      console.log('User response:', response);
      setUser(response.user);
      console.log('Пользователь получен:', response.user);
    } catch (err) {
      console.error('Ошибка получения пользователя:', err);
      
      // Если это тестовый пользователь, попробуем создать его
      if (telegramId === 12345678) {
        console.log('Trying to create test user...');
        try {
          const createResponse = await ApiService.createUser({
            telegramId: telegramId,
            username: 'test_user',
            firstName: 'Test',
            lastName: 'User'
          });
          console.log('Test user created:', createResponse);
          setUser(createResponse.user);
          console.log('Тестовый пользователь создан:', createResponse.user);
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
      console.log('fetchQueueStatus called, isInitialLoad:', isInitialLoad);
      
      // Устанавливаем loading только при первой загрузке
      if (isInitialLoad) {
        setLoading(true);
      }
      
      console.log('Calling ApiService.getQueueStatus...');
      const data = await ApiService.getQueueStatus();
      console.log('Queue status data received:', data);
      console.log('All boxes:', data.all_boxes || data.allBoxes);
      console.log('Wash queue:', data.wash_queue || data.washQueue);
      console.log('Air dry queue:', data.air_dry_queue || data.airDryQueue);
      console.log('Vacuum queue:', data.vacuum_queue || data.vacuumQueue);
      
      // Обновляем данные, сохраняя структуру объекта
      setWashInfo(prevInfo => {
        console.log('Updating washInfo, prevInfo:', prevInfo);
        
        // Если это первая загрузка или предыдущих данных нет
        if (!prevInfo) {
          const newInfo = {
            allBoxes: data.all_boxes || data.allBoxes || [],
            washQueue: data.wash_queue || data.washQueue || { queueSize: 0, hasQueue: false },
            airDryQueue: data.air_dry_queue || data.airDryQueue || { queueSize: 0, hasQueue: false },
            vacuumQueue: data.vacuum_queue || data.vacuumQueue || { queueSize: 0, hasQueue: false },
            totalQueueSize: data.total_queue_size || data.totalQueueSize || 0,
            hasAnyQueue: data.has_any_queue || data.hasAnyQueue || false,
            userSession: data.user_session || data.userSession || null
          };
          console.log('Created new washInfo:', newInfo);
          return newInfo;
        }
        
        // Обновляем только изменившиеся данные
        const updatedInfo = {
          ...prevInfo,
          allBoxes: data.all_boxes || data.allBoxes || prevInfo.allBoxes,
          washQueue: data.wash_queue || data.washQueue || prevInfo.washQueue,
          airDryQueue: data.air_dry_queue || data.airDryQueue || prevInfo.airDryQueue,
          vacuumQueue: data.vacuum_queue || data.vacuumQueue || prevInfo.vacuumQueue,
          totalQueueSize: data.total_queue_size || data.totalQueueSize || prevInfo.totalQueueSize,
          hasAnyQueue: data.has_any_queue || data.hasAnyQueue || prevInfo.hasAnyQueue,
          userSession: data.user_session || data.userSession || prevInfo.userSession
        };
        console.log('Updated washInfo:', updatedInfo);
        return updatedInfo;
      });
      
      setError(null);
      console.log('fetchQueueStatus completed successfully');
    } catch (err) {
      console.error('Ошибка загрузки информации о мойке:', err);
      setError('Не удалось загрузить информацию о мойке');
      
      // Создаем пустой объект с необходимой структурой только если нет предыдущих данных
      if (!washInfo) {
        console.log('Creating fallback washInfo due to error');
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
        console.log('Setting loading to false');
        setLoading(false);
      }
    }
  };
  
  // Функция для создания сессии
  const handleCreateSession = async (serviceData) => {
    try {
      if (!user) {
        setError('Пользователь не авторизован');
        return;
      }
      
      setLoading(true);
      
      // Добавляем данные о типе услуги, химии и номере машины в запрос
      const response = await ApiService.createSession({ 
        user_id: user.id,
        service_type: serviceData.serviceType,
        with_chemistry: serviceData.withChemistry,
        car_number: serviceData.carNumber,
        rental_time_minutes: serviceData.rentalTimeMinutes
      });
      
      console.log('Создана сессия:', response);
      
      // Обновляем информацию о сессии в состоянии
      setWashInfo(prevInfo => ({
        ...prevInfo,
        userSession: response.session
      }));
      
      // Запускаем поллинг для обновления статуса сессии по session_id
      if (response.session && ['created', 'assigned', 'active'].includes(response.session.status)) {
        startSessionPolling(response.session.id);
      }
    } catch (err) {
      setError('Не удалось создать сессию');
      console.error('Ошибка создания сессии:', err);
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
        console.log('Обновление статуса сессии:', sessionData);
        
        if (sessionData && sessionData.session) {
          // Обновляем данные сессии пользователя
          setWashInfo(prevInfo => {
            console.log('Обновление сессии в washInfo:', sessionData.session);
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
    console.log('useEffect for polling started, user:', user);
    
    if (user) {
      try {
        console.log('User found, starting data loading...');
        
        // Загружаем статус очереди и боксов при первой загрузке
        fetchQueueStatus(true);
        
        // Запускаем поллинг статуса очереди (должен работать всегда)
        const queueInterval = startQueuePolling();

        // Загружаем информацию о сессии пользователя по user_id и запускаем поллинг по session_id
        const loadUserSessionAndStartPolling = async () => {
          try {
            console.log('Loading user session for user ID:', user.id);
            
            // Сначала пробуем получить сессию напрямую через getUserSession
            const sessionResponse = await ApiService.getUserSession(user.id);
            console.log('User session direct data:', sessionResponse);
            
            if (sessionResponse && sessionResponse.session) {
              // Обновляем данные сессии пользователя
              setWashInfo(prevInfo => {
                console.log('Updating userSession in washInfo:', sessionResponse.session);
                return {
                  ...prevInfo,
                  userSession: sessionResponse.session
                };
              });
              
              // Если у пользователя есть активная сессия, запускаем поллинг для неё по ID сессии
              if (['created', 'assigned', 'active'].includes(sessionResponse.session.status)) {
                console.log('Starting session polling for session ID:', sessionResponse.session.id);
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
          console.log('Cleaning up queue interval');
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

  const themeObject = getTheme(theme);

  console.log('TelegramApp render - loading:', loading, 'error:', error, 'user:', user, 'washInfo:', washInfo);

  return (
    <AppContainer theme={themeObject}>
      <Header theme={theme} />
        <ContentContainer>
          {/* Отладочная информация */}
          <div style={{ 
            background: '#f0f0f0', 
            padding: '10px', 
            margin: '10px 0', 
            fontSize: '12px', 
            fontFamily: 'monospace' 
          }}>
            <div>Loading: {loading ? 'true' : 'false'}</div>
            <div>Error: {error || 'none'}</div>
            <div>User: {user ? `ID: ${user.id}, Name: ${user.firstName || user.first_name || 'N/A'}` : 'none'}</div>
            <div>WashInfo: {washInfo ? `Boxes: ${(washInfo.allBoxes || washInfo.all_boxes || []).length}, Queues: ${washInfo.hasAnyQueue || washInfo.has_any_queue ? 'yes' : 'no'}` : 'none'}</div>
            <div>API URL: {process.env.REACT_APP_API_URL || 'http://localhost:8080/api'}</div>
          </div>
          
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
                      onCreateSession={handleCreateSession}
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
            <Route path="*" element={<Navigate to="/telegram" />} />
          </Routes>
        </ContentContainer>
      </AppContainer>
  );
};

export default TelegramApp;
