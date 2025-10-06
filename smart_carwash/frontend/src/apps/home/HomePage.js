import React from 'react';
import styled from 'styled-components';
import { Link } from 'react-router-dom';
import { getTheme } from '../../shared/styles/theme';

const HomeContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 20px;
  background-color: ${props => props.theme.backgroundColor};
  color: ${props => props.theme.textColor};
`;

const Title = styled.h1`
  font-size: 2.5rem;
  margin-bottom: 20px;
  text-align: center;
`;

const Description = styled.p`
  font-size: 1.2rem;
  margin-bottom: 30px;
  text-align: center;
  max-width: 800px;
`;

const ButtonContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-top: 20px;
  
  @media (min-width: 768px) {
    flex-direction: row;
  }
`;

const StyledButton = styled.a`
  display: inline-block;
  padding: 12px 24px;
  background-color: var(--tg-theme-button-color);
  color: #ffffff;
  border-radius: 8px;
  text-decoration: none;
  font-weight: bold;
  text-align: center;
  transition: opacity 0.3s ease;
  
  &:hover {
    opacity: 0.9;
    text-decoration: none;
  }
`;

const StyledLink = styled(Link)`
  display: inline-block;
  padding: 12px 24px;
  background-color: var(--tg-theme-button-color);
  color: #ffffff;
  border-radius: 8px;
  text-decoration: none;
  font-weight: bold;
  text-align: center;
  transition: opacity 0.3s ease;
  
  &:hover {
    opacity: 0.9;
    text-decoration: none;
  }
`;

const HomePage = () => {
  const theme = getTheme('light');

  return (
    <HomeContainer theme={theme}>
      <Title>Умная автомойка H2O</Title>
      <Description>
        Добро пожаловать в систему умной автомойки H2O! Наша система позволяет удобно управлять процессом мойки автомобиля, 
        отслеживать статус боксов и очереди, а также просматривать историю ваших сессий.
      </Description>
      
      <ButtonContainer>
        <StyledButton 
          href="https://t.me/carwash_grom_test_bot?start=webapp" 
          target="_blank" 
          rel="noopener noreferrer"
        >
          Открыть в Telegram
        </StyledButton>
        
        <StyledLink to="/cashier/login">
          Интерфейс кассира
        </StyledLink>
        
        <StyledLink to="/admin/login">
          Интерфейс администратора
        </StyledLink>

        <StyledLink to="/cleaner/login">
          Интерфейс уборщика
        </StyledLink>
      </ButtonContainer>
    </HomeContainer>
  );
};

export default HomePage;
