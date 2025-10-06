import React, { useState, useEffect, useRef } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../shared/styles/theme';
import AuthService from '../../shared/services/AuthService';
import ApiService from '../../shared/services/ApiService';
import WashBoxList from './components/WashBoxList';
import CleaningTimer from './components/CleaningTimer';

const CleanerContainer = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: ${props => props.theme.backgroundColor};
  color: ${props => props.theme.textColor};
`;

const Header = styled.header`
  background-color: ${props => props.theme.cardBackground};
  padding: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const Title = styled.h1`
  margin: 0;
  font-size: 1.5rem;
`;

const LogoutButton = styled.button`
  background: none;
  border: none;
  color: ${props => props.theme.textColor};
  cursor: pointer;
  font-size: 1rem;
  
  &:hover {
    color: ${props => props.theme.primaryColor};
  }
`;

const UserInfo = styled.div`
  display: flex;
  align-items: center;
`;

const Username = styled.span`
  margin-right: 15px;
  font-weight: 500;
`;

const Content = styled.main`
  flex: 1;
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
`;

const Card = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

const ErrorMessage = styled.div`
  color: #dc3545;
  padding: 16px;
  background-color: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 6px;
  margin-bottom: 16px;
`;

/**
 * Приложение уборщика
 * @returns {React.ReactNode} - Приложение уборщика
 */
const CleanerApp = () => {
  const theme = getTheme('light');
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const cleaningTimerRef = useRef(null);

  useEffect(() => {
    // Проверяем авторизацию при загрузке компонента
    const checkAuth = () => {
      const currentUser = AuthService.getCurrentUser();
      const isAuthenticated = AuthService.isAuthenticated();
      const isAdmin = AuthService.isAdmin();
      
      // Если пользователь не авторизован, перенаправляем на страницу входа
      if (!isAuthenticated) {
        window.location.href = '/cleaner/login';
        return;
      }
      
      // Если пользователь администратор, перенаправляем на страницу администратора
      if (isAdmin) {
        window.location.href = '/admin';
        return;
      }
      
      // Если все проверки пройдены, устанавливаем пользователя
      setUser(currentUser);
      setIsLoading(false);
    };
    
    checkAuth();
  }, []);

  // Обработчик выхода из системы
  const handleLogout = async () => {
    await AuthService.logout();
    window.location.href = '/cleaner/login';
  };

  // Функция для обновления таймера уборки
  const refreshCleaningTimer = () => {
    if (cleaningTimerRef.current) {
      cleaningTimerRef.current.refreshData();
    }
  };
  
  // Если идет загрузка, показываем пустой контент
  if (isLoading) {
    return null;
  }
  
  return (
    <CleanerContainer theme={theme}>
      <Header theme={theme}>
        <Title>Интерфейс уборщика</Title>
        {user && (
          <UserInfo>
            <Username>{user.username}</Username>
            <LogoutButton onClick={handleLogout} theme={theme}>
              Выйти
            </LogoutButton>
          </UserInfo>
        )}
      </Header>
      <Content>
        {error && (
          <ErrorMessage>{error}</ErrorMessage>
        )}

        <Card theme={theme}>
          <h2>Активная уборка</h2>
          <CleaningTimer ref={cleaningTimerRef} />
        </Card>

        <Card theme={theme}>
          <h2>Боксы мойки</h2>
          <WashBoxList onCleaningAction={refreshCleaningTimer} />
        </Card>
      </Content>
    </CleanerContainer>
  );
};

export default CleanerApp;

