import React from 'react';
import styled from 'styled-components';
import { Link } from 'react-router-dom';
import { getTheme } from '../../../shared/styles/theme';

const Container = styled.div`
  padding: 20px;
`;

const Title = styled.h1`
  margin: 0 0 30px 0;
  color: ${props => props.theme.textColor};
  font-size: 2rem;
`;

const Grid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
`;

const Card = styled(Link)`
  background-color: ${props => props.theme.cardBackground};
  padding: 25px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  text-decoration: none;
  color: ${props => props.theme.textColor};
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
    color: ${props => props.theme.textColor};
  }
`;

const CardTitle = styled.h3`
  margin: 0 0 10px 0;
  color: ${props => props.theme.primaryColor};
  font-size: 1.2rem;
`;

const CardDescription = styled.p`
  margin: 0;
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;
  line-height: 1.4;
`;

const Icon = styled.div`
  font-size: 2rem;
  margin-bottom: 15px;
  color: ${props => props.theme.primaryColor};
`;

const WelcomeSection = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 25px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 30px;
`;

const WelcomeTitle = styled.h2`
  margin: 0 0 15px 0;
  color: ${props => props.theme.textColor};
`;

const WelcomeText = styled.p`
  margin: 0;
  color: ${props => props.theme.textColor};
  line-height: 1.6;
`;

/**
 * Компонент панели управления администратора
 * @returns {React.ReactNode} - Панель управления
 */
const Dashboard = () => {
  const theme = getTheme('light');

  const sections = [
    {
      title: 'Боксы мойки',
      description: 'Управление боксами мойки: создание, редактирование, удаление и мониторинг статуса боксов.',
      icon: '🚗',
      path: '/admin/washboxes'
    },
    {
      title: 'Сессии мойки',
      description: 'Просмотр и управление сессиями мойки с фильтрацией по статусу, пользователю и дате.',
      icon: '⏱️',
      path: '/admin/sessions'
    },
    {
      title: 'Очередь',
      description: 'Мониторинг текущего состояния очереди и просмотр пользователей, ожидающих обслуживания.',
      icon: '📋',
      path: '/admin/queue'
    },
    {
      title: 'Пользователи',
      description: 'Управление пользователями системы: просмотр списка, информации о пользователях.',
      icon: '👥',
      path: '/admin/users'
    },
    {
      title: 'Управление кассирами',
      description: 'Создание и управление учетными записями кассиров для работы с системой.',
      icon: '👨‍💼',
      path: '/admin/cashiers'
    }
  ];

  return (
    <Container>
      <WelcomeSection theme={theme}>
        <WelcomeTitle theme={theme}>Добро пожаловать в панель администратора</WelcomeTitle>
        <WelcomeText theme={theme}>
          Здесь вы можете управлять всеми аспектами системы умной автомойки. 
          Выберите нужный раздел для выполнения административных задач.
        </WelcomeText>
      </WelcomeSection>

      <Title theme={theme}>Разделы управления</Title>

      <Grid>
        {sections.map((section, index) => (
          <Card key={index} to={section.path} theme={theme}>
            <Icon theme={theme}>{section.icon}</Icon>
            <CardTitle theme={theme}>{section.title}</CardTitle>
            <CardDescription theme={theme}>{section.description}</CardDescription>
          </Card>
        ))}
      </Grid>
    </Container>
  );
};

export default Dashboard;
