import React, { useEffect, useState } from 'react';
import WebApp from '@twa-dev/sdk';
import styled from 'styled-components';
import Header from './components/Header';
import WashInfo from './components/WashInfo';
import WelcomeMessage from './components/WelcomeMessage';
import ApiService from './services/ApiService';

const AppContainer = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: ${props => props.theme === 'dark' ? '#1E1E1E' : '#F5F5F5'};
  color: ${props => props.theme === 'dark' ? '#FFFFFF' : '#000000'};
  transition: background-color 0.3s ease, color 0.3s ease;
`;

const ContentContainer = styled.div`
  flex: 1;
  padding: 16px;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
`;

function App() {
  const [washInfo, setWashInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [theme, setTheme] = useState('light');
  const [user, setUser] = useState(null);

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
        
        // Создаем или получаем пользователя
        createUser({
          telegram_id: telegramUser.id,
          username: telegramUser.username || '',
          first_name: telegramUser.first_name || '',
          last_name: telegramUser.last_name || ''
        });
      } else {
        // Для разработки используем тестового пользователя
        createUser({
          telegram_id: 12345678,
          username: 'test_user',
          first_name: 'Test',
          last_name: 'User'
        });
      }
    } catch (err) {
      console.error('Ошибка инициализации Telegram WebApp:', err);
      console.log('Продолжаем работу в режиме разработки');
      
      // Для разработки используем тестового пользователя
      createUser({
        telegram_id: 12345678,
        username: 'test_user',
        first_name: 'Test',
        last_name: 'User'
      });
    }
  }, []);
  
  // Эффект для загрузки информации о мойке при изменении пользователя
  useEffect(() => {
    if (user) {
      fetchWashInfoForUser(true); // Первая загрузка
    }
  }, [user]);

  // Функция для создания пользователя
  const createUser = async (userData) => {
    try {
      const response = await ApiService.createUser(userData);
      setUser(response.user);
      console.log('Пользователь создан/получен:', response.user);
    } catch (err) {
      console.error('Ошибка создания пользователя:', err);
      setError('Не удалось создать пользователя');
    }
  };

  // Функция для загрузки информации о мойке для пользователя
  const fetchWashInfoForUser = async (isInitialLoad = false) => {
    try {
      // Устанавливаем loading только при первой загрузке
      if (isInitialLoad) {
        setLoading(true);
      }
      
      const data = await ApiService.getWashInfoForUser(user.id);
      console.log('API response data:', data);
      
      // Обновляем данные, сохраняя структуру объекта
      setWashInfo(prevInfo => {
        // Если это первая загрузка или предыдущих данных нет
        if (!prevInfo) {
          return data;
        }
        
        // Обновляем только изменившиеся данные
        return {
          ...prevInfo,
          boxes: data.boxes,
          queueSize: data.queueSize,
          hasQueue: data.hasQueue,
          userSession: data.user_session
        };
      });
      
      setError(null);
    } catch (err) {
      setError('Не удалось загрузить информацию о мойке');
      console.error('Ошибка загрузки информации о мойке:', err);
      
      // Создаем пустой объект с необходимой структурой только если нет предыдущих данных
      if (!washInfo) {
        setWashInfo({
          boxes: [],
          queueSize: 0,
          hasQueue: false,
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
  const handleCreateSession = async () => {
    try {
      if (!user) {
        setError('Пользователь не авторизован');
        return;
      }
      
      setLoading(true);
      await ApiService.createSession({ user_id: user.id });
      
      // Обновляем информацию о мойке после создания сессии
      await fetchWashInfoForUser(false);
      
      // Запускаем поллинг для обновления статуса сессии
      startPolling();
    } catch (err) {
      setError('Не удалось создать сессию');
      console.error('Ошибка создания сессии:', err);
    } finally {
      setLoading(false);
    }
  };
  
  // Функция для запуска поллинга
  const startPolling = () => {
    // Устанавливаем интервал для поллинга (каждые 5 секунд)
    const pollingInterval = setInterval(() => {
      if (user) {
        fetchWashInfoForUser(false); // Не первая загрузка, обновление данных
      } else {
        // Если пользователь не определен, останавливаем поллинг
        clearInterval(pollingInterval);
      }
    }, 5000);
    
    // Сохраняем интервал в ref, чтобы можно было его очистить при размонтировании компонента
    // Для простоты в этом примере мы не очищаем интервал
  };

  return (
    <AppContainer theme={theme}>
      <Header theme={theme} />
      <ContentContainer>
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
          />
        )}
      </ContentContainer>
    </AppContainer>
  );
}

export default App;
