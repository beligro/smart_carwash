import React from 'react';
import styled from 'styled-components';
import { getTheme } from '../../shared/styles/theme';
import LoginForm from '../../shared/components/LoginForm';
import AuthService from '../../shared/services/AuthService';

const LoginContainer = styled.div`
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
  justify-content: center;
  align-items: center;
`;

const Title = styled.h1`
  margin: 0;
  font-size: 1.5rem;
`;

const Content = styled.main`
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 20px;
`;

/**
 * Страница авторизации администратора
 * @returns {React.ReactNode} - Страница авторизации администратора
 */
const AdminLoginPage = () => {
  const theme = getTheme('light');

  // Функция для авторизации администратора
  const handleLogin = async (username, password) => {
    return AuthService.loginAdmin(username, password);
  };

  return (
    <LoginContainer theme={theme}>
      <Header theme={theme}>
        <Title>Умная автомойка - Вход для администратора</Title>
      </Header>
      <Content>
        <LoginForm 
          title="Вход для администратора" 
          onLogin={handleLogin} 
        />
      </Content>
    </LoginContainer>
  );
};

export default AdminLoginPage;
