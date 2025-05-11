import React, { useEffect, useState } from 'react';
import { useSDK } from '@tma.js/sdk-react';
import styled from 'styled-components';
import HomePage from './pages/HomePage';
import LoadingScreen from './components/LoadingScreen';
import ErrorScreen from './components/ErrorScreen';

const AppContainer = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
`;

function App() {
  const { didInit, components, error } = useSDK();
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    // Set the app as ready when SDK is initialized
    if (didInit) {
      setIsReady(true);
    }
  }, [didInit]);

  // Show loading screen while SDK is initializing
  if (!didInit) {
    return <LoadingScreen message="Инициализация приложения..." />;
  }

  // Show error screen if SDK initialization failed
  if (error) {
    return <ErrorScreen message="Ошибка инициализации приложения" error={error} />;
  }

  // Main app content
  return (
    <AppContainer>
      {components.BackButton && <components.BackButton />}
      <HomePage />
    </AppContainer>
  );
}

export default App;
