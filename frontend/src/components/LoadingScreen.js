import React from 'react';
import styled from 'styled-components';

const LoadingContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background-color: var(--tg-theme-bg-color, #ffffff);
  color: var(--tg-theme-text-color, #000000);
`;

const Spinner = styled.div`
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  border-top: 4px solid var(--tg-theme-button-color, #007bff);
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;

const Message = styled.p`
  font-size: 16px;
  text-align: center;
  margin: 0;
  padding: 0 20px;
`;

const LoadingScreen = ({ message = 'Загрузка...' }) => {
  return (
    <LoadingContainer>
      <Spinner />
      <Message>{message}</Message>
    </LoadingContainer>
  );
};

export default LoadingScreen;
