import React from 'react';
import styled from 'styled-components';

const ErrorContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
  padding: 0 20px;
`;

const ErrorIcon = styled.div`
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background-color: var(--danger-color, #dc3545);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
  
  &::before {
    content: '!';
    font-size: 40px;
    font-weight: bold;
    color: white;
  }
`;

const ErrorTitle = styled.h2`
  font-size: 20px;
  text-align: center;
  margin: 0 0 10px 0;
`;

const ErrorMessage = styled.p`
  font-size: 16px;
  text-align: center;
  margin: 0 0 20px 0;
  color: var(--tg-theme-hint-color, #6c757d);
`;

const ErrorDetails = styled.div`
  background-color: var(--tg-theme-secondary-bg-color, #f8f9fa);
  padding: 15px;
  border-radius: 8px;
  width: 100%;
  max-width: 400px;
  margin-top: 10px;
  font-family: monospace;
  font-size: 14px;
  overflow-x: auto;
`;

const RefreshButton = styled.button`
  margin-top: 20px;
`;

const ErrorScreen = ({ message = 'Произошла ошибка', error }) => {
  const handleRefresh = () => {
    window.location.reload();
  };

  return (
    <ErrorContainer>
      <ErrorIcon />
      <ErrorTitle>Ошибка</ErrorTitle>
      <ErrorMessage>{message}</ErrorMessage>
      
      {error && (
        <ErrorDetails>
          {error.toString()}
        </ErrorDetails>
      )}
      
      <RefreshButton onClick={handleRefresh}>
        Обновить страницу
      </RefreshButton>
    </ErrorContainer>
  );
};

export default ErrorScreen;
