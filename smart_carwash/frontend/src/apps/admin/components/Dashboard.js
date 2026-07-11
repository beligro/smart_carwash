import React from 'react';
import styled from 'styled-components';
import { Link } from 'react-router-dom';
import { getTheme } from '../../../shared/styles/theme';
import AuthService from '../../../shared/services/AuthService';

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

  const isLimitedAdmin = AuthService.getRole() === 'limited_admin';

  const sections = [
    {
      title: 'Боксы мойки',
      description: 'Управление боксами мойки: создание, редактирование, удаление и мониторинг статуса боксов.',
      icon: '🚗',
      path: '/admin/washboxes',
      section: 'washboxes'
    },
    !isLimitedAdmin ? {
      title: 'История боксов',
      description: 'Полная история изменений статусов, света и химии по всем боксам.',
      icon: '🗂️',
      path: '/admin/washbox-change-logs',
      section: 'washbox-change-logs'
    } : null,
    {
      title: 'Сессии мойки',
      description: 'Просмотр и управление сессиями мойки с фильтрацией по статусу, пользователю и дате.',
      icon: '⏱️',
      path: '/admin/sessions',
      section: 'sessions'
    },
    {
      title: 'Очередь',
      description: 'Мониторинг текущего состояния очереди и просмотр клиентов, ожидающих обслуживания.',
      icon: '📋',
      path: '/admin/queue',
      section: 'queue'
    },
    {
      title: 'Клиенты',
      description: 'Управление клиентами системы: просмотр списка, информации о клиентах.',
      icon: '👥',
      path: '/admin/users',
      section: 'users'
    },
    {
      title: 'Управление кассирами',
      description: 'Создание и управление учетными записями кассиров для работы с системой.',
      icon: '👨‍💼',
      path: '/admin/cashiers',
      section: 'cashiers'
    },
    {
      title: 'Управление уборщиками',
      description: 'Создание и управление учетными записями уборщиков для обслуживания боксов.',
      icon: '🧹',
      path: '/admin/cleaners',
      section: 'cleaners'
    },
    {
      title: 'Логи уборки',
      description: 'Просмотр и анализ логов уборки боксов с фильтрацией по уборщику, боксу и времени.',
      icon: '📊',
      path: '/admin/cleaning-logs',
      section: 'cleaning-logs'
    },
    {
      title: 'Платежи',
      description: 'Управление платежами и возвратами: просмотр, фильтрация и обработка платежей.',
      icon: '💳',
      path: '/admin/payments',
      section: 'payments'
    },
    {
      title: 'Настройки',
      description: 'Управление ценами и настройками услуг: изменение цен и времени мойки.',
      icon: '⚙️',
      path: '/admin/settings',
      section: 'settings'
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
        {sections
          .filter(Boolean)
          .filter((s) => AuthService.canAccessSection(s.section))
          .map((section, index) => (
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
