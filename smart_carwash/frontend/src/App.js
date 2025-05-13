import React, { useEffect, useState } from 'react';
import { WebApp } from '@twa-dev/sdk';
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

  useEffect(() => {
    // Инициализация Telegram Mini App
    WebApp.ready();
    
    // Установка темы в соответствии с темой Telegram
    const colorScheme = WebApp.colorScheme;
    setTheme(colorScheme);
    
    // Настройка основного цвета
    document.documentElement.style.setProperty(
      '--tg-theme-button-color', 
      WebApp.themeParams.button_color || '#2481cc'
    );
    
    // Настройка цвета текста
    document.documentElement.style.setProperty(
      '--tg-theme-text-color', 
      WebApp.themeParams.text_color || '#000000'
    );
    
    // Загрузка информации о мойке
    fetchWashInfo();
  }, []);

  // Функция для загрузки информации о мойке
  const fetchWashInfo = async () => {
    try {
      setLoading(true);
      const data = await ApiService.getWashInfo();
      setWashInfo(data);
      setError(null);
    } catch (err) {
      setError('Не удалось загрузить информацию о мойке');
      console.error('Ошибка загрузки информации о мойке:', err);
    } finally {
      setLoading(false);
    }
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
          <WashInfo washInfo={washInfo} theme={theme} />
        )}
      </ContentContainer>
    </AppContainer>
  );
}

export default App;
