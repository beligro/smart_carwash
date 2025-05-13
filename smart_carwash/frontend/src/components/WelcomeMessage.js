import React from 'react';
import styled from 'styled-components';
import { WebApp } from '@twa-dev/sdk';

const WelcomeContainer = styled.div`
  margin-bottom: 24px;
  padding: 20px;
  border-radius: 12px;
  background-color: ${props => props.theme === 'dark' ? '#2C2C2C' : '#FFFFFF'};
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: background-color 0.3s ease;
`;

const Title = styled.h1`
  font-size: 24px;
  margin-bottom: 16px;
  color: ${props => props.theme === 'dark' ? '#FFFFFF' : '#000000'};
`;

const Description = styled.p`
  font-size: 16px;
  line-height: 1.5;
  margin-bottom: 16px;
  color: ${props => props.theme === 'dark' ? '#E0E0E0' : '#333333'};
`;

const WelcomeMessage = ({ theme }) => {
  // Получаем имя пользователя из Telegram WebApp
  const firstName = WebApp.initDataUnsafe?.user?.first_name || 'Гость';

  return (
    <WelcomeContainer theme={theme}>
      <Title theme={theme}>Привет, {firstName}! 👋</Title>
      <Description theme={theme}>
        Добро пожаловать в приложение умной автомойки. Здесь вы можете увидеть 
        информацию о доступных боксах и их текущем статусе.
      </Description>
      <Description theme={theme}>
        Наша автомойка оборудована современными системами, которые обеспечивают 
        качественную и быструю мойку вашего автомобиля.
      </Description>
    </WelcomeContainer>
  );
};

export default WelcomeMessage;
