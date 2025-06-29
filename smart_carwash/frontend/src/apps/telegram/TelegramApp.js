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
      
      console.log('Telegram WebApp инициализирован успешно');
      
      // Получаем данные пользователя из Telegram
      if (WebApp.initDataUnsafe && WebApp.initDataUnsafe.user) {
        const telegramUser = WebApp.initDataUnsafe.user;
        
        // Получаем пользователя по telegram_id
        getUserByTelegramId(telegramUser.id);
      } else {
        // Для разработки используем тестового пользователя
        getUserByTelegramId(12345678);
      }
    } catch (err) {
      console.error('Ошибка инициализации Telegram WebApp:', err);
      console.log('Продолжаем работу в режиме разработки');
      
      // Для разработки используем тестового пользователя
      getUserByTelegramId(12345678);
    }
  }, []);
  
  // Функция для получения пользователя по telegram_id
  const getUserByTelegramId = async (telegramId) => {
    try {
      const response = await ApiService.getUserByTelegramId(telegramId);
      setUser(response.user);
      console.log('Пользователь получен:', response.user);
    } catch (err) {
      console.error('Ошибка получения пользователя:', err);
      setError('Не удалось получить пользователя. Возможно, вы не зарегистрированы в боте. Пожалуйста, нажмите /start в боте.');
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
      console.log('Queue status data:', data);
      
      // Обновляем данные, сохраняя структуру объекта
      setWashInfo(prevInfo => {
        // Если это первая загрузка или предыдущих данных нет
        if (!prevInfo) {
          return {
            allBoxes: data.all_boxes || [],
            washQueue: data.wash_queue || { queue_size: 0, has_queue: false },
            airDryQueue: data.air_dry_queue || { queue_size: 0, has_queue: false },
            vacuumQueue: data.vacuum_queue || { queue_size: 0, has_queue: false },
            totalQueueSize: data.total_queue_size || 0,
            hasAnyQueue: data.has_any_queue || false,
            userSession: data.user_session || null
          };
        }
        
        // Обновляем только изменившиеся данные
        return {
          ...prevInfo,
          allBoxes: data.all_boxes || prevInfo.allBoxes,
          washQueue: data.wash_queue || prevInfo.washQueue,
          airDryQueue: data.air_dry_queue || prevInfo.airDryQueue,
          vacuumQueue: data.vacuum_queue || prevInfo.vacuumQueue,
          totalQueueSize: data.total_queue_size || prevInfo.totalQueueSize,
          hasAnyQueue: data.has_any_queue || prevInfo.hasAnyQueue,
          userSession: data.user_session || prevInfo.userSession
        };
      });
      
      setError(null);
    } catch (err) {
      setError('Не удалось загрузить информацию о мойке');
      console.error('Ошибка загрузки информации о мойке:', err);
      
      // Создаем пустой объект с необходимой структурой только если нет предыдущих данных
      if (!washInfo) {
        setWashInfo({
          allBoxes: [],
          washQueue: { serviceType: 'wash', boxes: [], queue_size: 0, has_queue: false },
          airDryQueue: { serviceType: 'air_dry', boxes: [], queue_size: 0, has_queue: false },
          vacuumQueue: { serviceType: 'vacuum', boxes: [], queue_size: 0, has_queue: false },
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
    if (user) {
      // Загружаем статус очереди и боксов при первой загрузке
      fetchQueueStatus(true);
      
      // Запускаем поллинг статуса очереди (должен работать всегда)
      const queueInterval = startQueuePolling();

      // Загружаем информацию о сессии пользователя по user_id и запускаем поллинг по session_id
      const loadUserSessionAndStartPolling = async () => {
        try {
          // Сначала пробуем получить сессию напрямую через getUserSession
          const sessionResponse = await ApiService.getUserSession(user.id);
          console.log('User session direct data:', sessionResponse);
          
          if (sessionResponse && sessionResponse.session) {
            // Обновляем данные сессии пользователя
            setWashInfo(prevInfo => ({
              ...prevInfo,
              userSession: sessionResponse.session
            }));
            
            // Если у пользователя есть активная сессия, запускаем поллинг для неё по ID сессии
            if (['created', 'assigned', 'active'].includes(sessionResponse.session.status)) {
              startSessionPolling(sessionResponse.session.id);
            }
          }
        } catch (err) {
          console.error('Ошибка загрузки информации о пользовательской сессии:', err);
        }
      };
      
      loadUserSessionAndStartPolling();
      
      // Очищаем интервал при размонтировании
      return () => clearInterval(queueInterval);
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
                  ) : (
                    <WashInfo 
                      washInfo={washInfo} 
                      theme={theme} 
                      onCreateSession={handleCreateSession}
                      onViewHistory={handleViewHistory}
                      user={user}
                    />
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
