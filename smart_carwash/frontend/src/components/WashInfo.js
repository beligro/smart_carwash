import React from 'react';
import styled from 'styled-components';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

const WashInfoContainer = styled.div`
  margin-bottom: 24px;
`;

const Section = styled.div`
  margin-bottom: 24px;
`;

const Title = styled.h2`
  font-size: 20px;
  margin-bottom: 16px;
  color: ${props => props.theme === 'dark' ? '#FFFFFF' : '#000000'};
`;

const Subtitle = styled.h3`
  font-size: 18px;
  margin-bottom: 12px;
  color: ${props => props.theme === 'dark' ? '#E0E0E0' : '#333333'};
`;

const BoxesGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
`;

const BoxCard = styled.div`
  padding: 16px;
  border-radius: 12px;
  background-color: ${props => props.theme === 'dark' ? '#2C2C2C' : '#FFFFFF'};
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s ease, background-color 0.3s ease;
  
  &:hover {
    transform: translateY(-4px);
  }
`;

const BoxNumber = styled.h3`
  font-size: 18px;
  margin-bottom: 8px;
  color: ${props => props.theme === 'dark' ? '#FFFFFF' : '#000000'};
`;

const BoxStatus = styled.div`
  display: inline-block;
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 500;
  background-color: ${props => {
    if (props.status === 'free') return props.theme === 'dark' ? '#2E7D32' : '#E8F5E9';
    if (props.status === 'busy') return props.theme === 'dark' ? '#C62828' : '#FFEBEE';
    return props.theme === 'dark' ? '#F57C00' : '#FFF3E0';
  }};
  color: ${props => {
    if (props.status === 'free') return props.theme === 'dark' ? '#FFFFFF' : '#2E7D32';
    if (props.status === 'busy') return props.theme === 'dark' ? '#FFFFFF' : '#C62828';
    return props.theme === 'dark' ? '#FFFFFF' : '#F57C00';
  }};
`;

const NoBoxesMessage = styled.p`
  font-size: 16px;
  color: ${props => props.theme === 'dark' ? '#E0E0E0' : '#666666'};
  text-align: center;
  padding: 24px;
  background-color: ${props => props.theme === 'dark' ? '#2C2C2C' : '#FFFFFF'};
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
`;

const InfoCard = styled.div`
  padding: 16px;
  border-radius: 12px;
  background-color: ${props => props.theme === 'dark' ? '#2C2C2C' : '#FFFFFF'};
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 16px;
`;

const QueueInfo = styled.div`
  display: flex;
  align-items: center;
  margin-bottom: 8px;
`;

const QueueStatus = styled.span`
  display: inline-block;
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 500;
  margin-left: 8px;
  background-color: ${props => props.hasQueue 
    ? (props.theme === 'dark' ? '#C62828' : '#FFEBEE') 
    : (props.theme === 'dark' ? '#2E7D32' : '#E8F5E9')};
  color: ${props => props.hasQueue 
    ? (props.theme === 'dark' ? '#FFFFFF' : '#C62828') 
    : (props.theme === 'dark' ? '#FFFFFF' : '#2E7D32')};
`;

const SessionCard = styled.div`
  padding: 16px;
  border-radius: 12px;
  background-color: ${props => props.theme === 'dark' ? '#2C2C2C' : '#FFFFFF'};
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 16px;
`;

const StatusIndicator = styled.div`
  display: flex;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid ${props => props.theme === 'dark' ? '#444444' : '#EEEEEE'};
`;

const StatusDot = styled.div`
  width: 10px;
  height: 10px;
  border-radius: 50%;
  margin-right: 8px;
  background-color: ${props => {
    switch(props.status) {
      case 'created': return '#FFC107'; // Желтый - в очереди
      case 'assigned': return '#2196F3'; // Синий - назначен бокс
      case 'active': return '#4CAF50'; // Зеленый - активна
      case 'complete': return '#8BC34A'; // Светло-зеленый - завершена
      case 'canceled': return '#F44336'; // Красный - отменена
      default: return '#9E9E9E'; // Серый - неизвестно
    }
  }};
  animation: ${props => props.status === 'created' ? 'pulse 1.5s infinite' : 'none'};
  
  @keyframes pulse {
    0% {
      transform: scale(0.95);
      box-shadow: 0 0 0 0 rgba(255, 193, 7, 0.7);
    }
    
    70% {
      transform: scale(1);
      box-shadow: 0 0 0 5px rgba(255, 193, 7, 0);
    }
    
    100% {
      transform: scale(0.95);
      box-shadow: 0 0 0 0 rgba(255, 193, 7, 0);
    }
  }
`;

const StatusText = styled.span`
  font-size: 14px;
  color: ${props => props.theme === 'dark' ? '#E0E0E0' : '#666666'};
`;

const SessionStatus = styled.div`
  display: inline-block;
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  background-color: ${props => {
    switch(props.status) {
      case 'created': return props.theme === 'dark' ? '#0D47A1' : '#E3F2FD';
      case 'assigned': return props.theme === 'dark' ? '#1565C0' : '#BBDEFB';
      case 'active': return props.theme === 'dark' ? '#2E7D32' : '#E8F5E9';
      case 'complete': return props.theme === 'dark' ? '#558B2F' : '#F1F8E9';
      case 'canceled': return props.theme === 'dark' ? '#C62828' : '#FFEBEE';
      default: return props.theme === 'dark' ? '#424242' : '#F5F5F5';
    }
  }};
  color: ${props => {
    switch(props.status) {
      case 'created': return props.theme === 'dark' ? '#FFFFFF' : '#0D47A1';
      case 'assigned': return props.theme === 'dark' ? '#FFFFFF' : '#1565C0';
      case 'active': return props.theme === 'dark' ? '#FFFFFF' : '#2E7D32';
      case 'complete': return props.theme === 'dark' ? '#FFFFFF' : '#558B2F';
      case 'canceled': return props.theme === 'dark' ? '#FFFFFF' : '#C62828';
      default: return props.theme === 'dark' ? '#FFFFFF' : '#424242';
    }
  }};
`;

const SessionInfo = styled.p`
  font-size: 14px;
  margin-bottom: 8px;
  color: ${props => props.theme === 'dark' ? '#E0E0E0' : '#666666'};
`;

const Button = styled.button`
  background-color: var(--tg-theme-button-color, #2481cc);
  color: #FFFFFF;
  border: none;
  border-radius: 8px;
  padding: 10px 16px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s ease;
  margin-top: 8px;
  
  &:hover {
    opacity: 0.9;
  }
  
  &:disabled {
    background-color: ${props => props.theme === 'dark' ? '#424242' : '#E0E0E0'};
    color: ${props => props.theme === 'dark' ? '#757575' : '#9E9E9E'};
    cursor: not-allowed;
    opacity: 0.7;
  }
`;

const getBoxStatusText = (status) => {
  switch (status) {
    case 'free':
      return 'Свободен';
    case 'busy':
      return 'Занят';
    case 'maintenance':
      return 'На обслуживании';
    default:
      return 'Неизвестно';
  }
};

const getSessionStatusText = (status) => {
  switch (status) {
    case 'created':
      return 'В очереди';
    case 'assigned':
      return 'Назначена';
    case 'active':
      return 'Активна';
    case 'complete':
      return 'Завершена';
    case 'canceled':
      return 'Отменена';
    default:
      return 'Неизвестно';
  }
};

const formatDate = (dateString) => {
  try {
    const date = new Date(dateString);
    return format(date, 'dd MMMM yyyy, HH:mm', { locale: ru });
  } catch (error) {
    console.error('Ошибка форматирования даты:', error);
    return dateString;
  }
};

const WashInfo = ({ washInfo, theme, onCreateSession }) => {
  const boxes = washInfo?.boxes || [];
  const queueSize = washInfo?.queueSize || 0;
  const hasQueue = washInfo?.hasQueue || false;
  
  // Отладочная информация
  console.log('WashInfo props:', { washInfo });
  console.log('User session from API:', washInfo?.user_session);
  
  // Не переопределяем userSession, а используем напрямую из washInfo
  const userSession = washInfo?.user_session;

  return (
    <WashInfoContainer>
      {/* Информация об очереди */}
      <Section>
        <Title theme={theme}>Статус автомойки</Title>
        <InfoCard theme={theme}>
          <QueueInfo>
            <span>Статус очереди:</span>
            <QueueStatus hasQueue={hasQueue} theme={theme}>
              {hasQueue ? 'Есть очередь' : 'Нет очереди'}
            </QueueStatus>
          </QueueInfo>
          {hasQueue && (
            <SessionInfo theme={theme}>
              Количество человек в очереди: {queueSize}
            </SessionInfo>
          )}
        </InfoCard>
      </Section>

      {/* Информация о сессии пользователя */}
      <Section>
        <Title theme={theme}>Ваша запись</Title>
        {userSession ? (
          <SessionCard theme={theme}>
            <SessionStatus status={userSession.status} theme={theme}>
              {getSessionStatusText(userSession.status)}
            </SessionStatus>
            <SessionInfo theme={theme}>
              Создана: {formatDate(userSession.created_at)}
            </SessionInfo>
            {userSession.box_id && (
              <SessionInfo theme={theme}>
                Назначен бокс: #{userSession.box_id}
              </SessionInfo>
            )}
            <StatusIndicator theme={theme}>
              <StatusDot status={userSession.status} />
              <StatusText theme={theme}>
                {userSession.status === 'created' 
                  ? 'Ожидание назначения бокса...' 
                  : userSession.status === 'assigned'
                    ? 'Бокс назначен, ожидание клиента'
                    : userSession.status === 'active'
                      ? 'Мойка в процессе'
                      : userSession.status === 'complete'
                        ? 'Мойка завершена'
                        : 'Сессия отменена'}
              </StatusText>
            </StatusIndicator>
          </SessionCard>
        ) : (
          <InfoCard theme={theme}>
            <SessionInfo theme={theme}>
              У вас нет активной записи на мойку
            </SessionInfo>
            <Button 
              theme={theme} 
              onClick={onCreateSession}
              disabled={hasQueue && queueSize > 5}
            >
              Записаться на мойку
            </Button>
            {hasQueue && queueSize > 5 && (
              <SessionInfo theme={theme} style={{ marginTop: '8px', color: '#C62828' }}>
                Очередь слишком большая, попробуйте позже
              </SessionInfo>
            )}
          </InfoCard>
        )}
      </Section>

      {/* Информация о боксах */}
      <Section>
        <Title theme={theme}>Боксы автомойки</Title>
        {boxes.length > 0 ? (
          <BoxesGrid>
            {boxes.map((box) => (
              <BoxCard key={box.id} theme={theme}>
                <BoxNumber theme={theme}>Бокс #{box.number}</BoxNumber>
                <BoxStatus status={box.status} theme={theme}>
                  {getBoxStatusText(box.status)}
                </BoxStatus>
              </BoxCard>
            ))}
          </BoxesGrid>
        ) : (
          <NoBoxesMessage theme={theme}>
            Информация о боксах отсутствует
          </NoBoxesMessage>
        )}
      </Section>
    </WashInfoContainer>
  );
};

export default WashInfo;
