import React from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';

const DashboardContainer = styled.div`
  display: flex;
  flex-direction: column;
`;

const Card = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

const Title = styled.h2`
  margin-top: 0;
  margin-bottom: 15px;
  color: ${props => props.theme.textColor};
`;

const Text = styled.p`
  color: ${props => props.theme.textColor};
  line-height: 1.5;
`;

/**
 * Компонент панели управления администратора
 * @returns {React.ReactNode} - Панель управления
 */
const Dashboard = () => {
  const theme = getTheme('light');

  return (
    <DashboardContainer>
      <Card theme={theme}>
        <Title theme={theme}>Панель управления</Title>
        <Text theme={theme}>
          Добро пожаловать в интерфейс администратора умной автомойки. Здесь вы можете управлять
          всеми аспектами системы, включая управление кассирами, настройку параметров и мониторинг работы.
        </Text>
      </Card>
      
      <Card theme={theme}>
        <Title theme={theme}>Возможности администратора</Title>
        <Text theme={theme}>
          <strong>Управление кассирами:</strong> Создание, редактирование и удаление учетных записей кассиров.
        </Text>
        <Text theme={theme}>
          <strong>Мониторинг системы:</strong> Отслеживание статуса боксов, очередей и активных сессий.
        </Text>
        <Text theme={theme}>
          <strong>Настройка параметров:</strong> Изменение настроек системы, таких как время аренды, стоимость услуг и т.д.
        </Text>
      </Card>
    </DashboardContainer>
  );
};

export default Dashboard;
