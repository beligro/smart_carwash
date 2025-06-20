import React from 'react';
import styled from 'styled-components';
import { getTheme } from '../../shared/styles/theme';

const AdminContainer = styled.div`
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

const AdminApp = () => {
  const theme = getTheme('light');

  return (
    <AdminContainer theme={theme}>
      <Header theme={theme}>
        <Title>Интерфейс администратора</Title>
      </Header>
      <Content>
        <Card theme={theme}>
          <h2>Добро пожаловать в интерфейс администратора</h2>
          <p>
            Этот интерфейс предназначен для администрирования системы умной автомойки. 
            Здесь будет реализован функционал для управления боксами, настройки параметров системы, 
            просмотра статистики и аналитики, а также другие административные функции.
          </p>
          <p>
            В настоящее время интерфейс находится в разработке. 
            Скоро здесь появятся все необходимые функции.
          </p>
        </Card>
      </Content>
    </AdminContainer>
  );
};

export default AdminApp;
