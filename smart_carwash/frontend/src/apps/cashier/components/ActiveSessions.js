import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import Timer from '../../../shared/components/UI/Timer';
import useTimer from '../../../shared/hooks/useTimer';

const Container = styled.div`
  padding: 20px;
`;

const SessionCard = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border-left: 4px solid ${props => {
    switch (props.status) {
      case 'in_queue': return '#17a2b8';
      case 'active': return '#28a745';
      case 'assigned': return '#007bff';
      case 'created': return '#ffc107';
      default: return '#6c757d';
    }
  }};
`;

const SessionHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
`;

const SessionInfo = styled.div`
  flex: 1;
`;

const SessionId = styled.div`
  font-weight: 600;
  font-size: 1.1rem;
  color: ${props => props.theme.textColor};
`;

const SessionStatus = styled.div`
  font-size: 0.9rem;
  color: ${props => props.theme.textColorSecondary};
  margin-top: 4px;
`;

const TimerContainer = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`;

const TimerLabel = styled.span`
  font-size: 0.9rem;
  color: ${props => props.theme.textColorSecondary};
`;

const SessionDetails = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
`;

const DetailItem = styled.div`
  display: flex;
  flex-direction: column;
`;

const DetailLabel = styled.span`
  font-size: 0.8rem;
  color: ${props => props.theme.textColorSecondary};
  margin-bottom: 4px;
`;

const DetailValue = styled.span`
  font-size: 0.9rem;
  color: ${props => props.theme.textColor};
  font-weight: 500;
`;

const ActionButtons = styled.div`
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
`;

const ActionButton = styled.button`
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &.start {
    background-color: #28a745;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #218838;
    }
  }

  &.complete {
    background-color: #dc3545;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #c82333;
    }
  }

  &.cancel {
    background-color: #6c757d;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #5a6268;
    }
  }
`;

const LoadingSpinner = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px;
  font-size: 1.1rem;
  color: ${props => props.theme.textColorSecondary};
`;

const ErrorMessage = styled.div`
  color: #dc3545;
  padding: 16px;
  background-color: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 6px;
  margin-bottom: 16px;
`;

const EmptyMessage = styled.div`
  text-align: center;
  padding: 40px;
  color: ${props => props.theme.textColorSecondary};
  font-size: 1.1rem;
`;

const formatDateTime = (dateString) => {
  return new Date(dateString).toLocaleString('ru-RU');
};

const formatAmount = (amount) => {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: 0
  }).format(amount / 100);
};

const getStatusText = (status) => {
  switch (status) {
    case 'created': return 'Создана';
    case 'in_queue': return 'В очереди';
    case 'assigned': return 'Назначена';
    case 'active': return 'Активна';
    default: return status;
  }
};

const getServiceTypeText = (serviceType) => {
  switch (serviceType) {
    case 'wash': return 'Мойка';
    case 'air_dry': return 'Сушка';
    case 'vacuum': return 'Пылесос';
    default: return serviceType;
  }
};

// Компонент кнопки включения химии для кассира
const ChemistryEnableButton = ({ session, onEnable, actionLoading, theme }) => {
  const [timeLeft, setTimeLeft] = useState(null);
  const [isExpired, setIsExpired] = useState(false);

  // Проверяем время доступности кнопки (по умолчанию 10 минут)
  useEffect(() => {
    if (!session || !session.status_updated_at) return;

    const checkTimeLimit = () => {
      const startTime = new Date(session.status_updated_at);
      const now = new Date();
      const timeLimit = 10 * 60 * 1000; // 10 минут в миллисекундах
      const timePassed = now - startTime;
      const remaining = timeLimit - timePassed;

      if (remaining <= 0) {
        setIsExpired(true);
        setTimeLeft(0);
      } else {
        setTimeLeft(Math.floor(remaining / 1000));
      }
    };

    checkTimeLimit();
    const interval = setInterval(checkTimeLimit, 1000);

    return () => clearInterval(interval);
  }, [session]);

  if (isExpired) {
    return null; // Не показываем ничего, когда время истекло
  }

  return (
    <ActionButton
      className="chemistry"
      onClick={() => onEnable(session.id)}
      disabled={actionLoading[session.id]}
      style={{ backgroundColor: '#4CAF50', color: 'white' }}
    >
      {actionLoading[session.id] ? 'Включение...' : '🧪 Включить химию'}
      {timeLeft !== null && (
        <span style={{ fontSize: '0.8rem', marginLeft: '4px' }}>
          ({Math.floor(timeLeft / 60)}:{(timeLeft % 60).toString().padStart(2, '0')})
        </span>
      )}
    </ActionButton>
  );
};

/**
 * Компонент для отображения одной сессии
 */
const SessionCardComponent = ({ session, onStart, onComplete, onCancel, onEnableChemistry, actionLoading, theme }) => {
  const { timeLeft } = useTimer(session);

  return (
    <SessionCard theme={theme} status={session.status}>
      <SessionHeader>
        <SessionInfo>
          <SessionId theme={theme}>Сессия #{session.id}</SessionId>
          <SessionStatus theme={theme}>
            {getStatusText(session.status)}
          </SessionStatus>
        </SessionInfo>
        
        {(session.status === 'active' || session.status === 'assigned') && timeLeft !== null && (
          <TimerContainer>
            <TimerLabel theme={theme}>Осталось:</TimerLabel>
            <Timer seconds={timeLeft} theme="light" />
          </TimerContainer>
        )}
      </SessionHeader>

      <SessionDetails>
        <DetailItem>
          <DetailLabel theme={theme}>Тип услуги</DetailLabel>
          <DetailValue theme={theme}>
            {getServiceTypeText(session.service_type)}
            {session.with_chemistry && ' + химия'}
          </DetailValue>
        </DetailItem>

        <DetailItem>
          <DetailLabel theme={theme}>Номер машины</DetailLabel>
          <DetailValue theme={theme}>
            {session.car_number || 'Не указан'}
          </DetailValue>
        </DetailItem>

        <DetailItem>
          <DetailLabel theme={theme}>Время аренды</DetailLabel>
          <DetailValue theme={theme}>
            {session.rental_time_minutes} мин
            {session.extension_time_minutes > 0 && ` + ${session.extension_time_minutes} мин`}
          </DetailValue>
        </DetailItem>

        <DetailItem>
          <DetailLabel theme={theme}>Создана</DetailLabel>
          <DetailValue theme={theme}>
            {formatDateTime(session.created_at)}
          </DetailValue>
        </DetailItem>

        {session.main_payment && (
          <DetailItem>
            <DetailLabel theme={theme}>Сумма</DetailLabel>
            <DetailValue theme={theme}>
              {formatAmount(session.main_payment.amount)}
            </DetailValue>
          </DetailItem>
        )}
      </SessionDetails>

      <ActionButtons>
        {session.status === 'assigned' && (
          <ActionButton
            className="start"
            onClick={() => onStart(session.id)}
            disabled={actionLoading[session.id]}
          >
            {actionLoading[session.id] ? 'Запускаем...' : 'Запустить'}
          </ActionButton>
        )}

        {session.status === 'active' && (
          <ActionButton
            className="complete"
            onClick={() => onComplete(session.id)}
            disabled={actionLoading[session.id]}
          >
            {actionLoading[session.id] ? 'Завершаем...' : 'Завершить'}
          </ActionButton>
        )}

        {(session.status === 'in_queue' || session.status === 'assigned') && (
          <ActionButton
            className="cancel"
            onClick={() => onCancel(session.id)}
            disabled={actionLoading[session.id]}
          >
            {actionLoading[session.id] ? 'Отменяем...' : 'Отменить'}
          </ActionButton>
        )}

        {/* Кнопка включения химии */}
        {session.status === 'active' && 
         session.with_chemistry && 
         !session.was_chemistry_on && (
          <ChemistryEnableButton
            session={session}
            onEnable={onEnableChemistry}
            actionLoading={actionLoading}
            theme={theme}
          />
        )}
      </ActionButtons>
    </SessionCard>
  );
};

/**
 * Компонент для отображения активных сессий кассира
 */
const ActiveSessions = () => {
  const theme = getTheme('light');
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState({});

  useEffect(() => {
    loadActiveSessions();
  }, []);

  const loadActiveSessions = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await ApiService.getCashierActiveSessions();
      setSessions(response.sessions || []);
    } catch (error) {
      console.error('Ошибка загрузки активных сессий:', error);
      setError('Ошибка загрузки активных сессий');
    } finally {
      setLoading(false);
    }
  };

  const handleStartSession = async (sessionId) => {
    setActionLoading(prev => ({ ...prev, [sessionId]: true }));
    
    try {
      await ApiService.startCashierSession(sessionId);
      await loadActiveSessions(); // Перезагружаем список
    } catch (error) {
      console.error('Ошибка запуска сессии:', error);
      setError('Ошибка запуска сессии');
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  const handleCompleteSession = async (sessionId) => {
    if (!window.confirm('Вы уверены, что хотите завершить эту сессию?')) {
      return;
    }

    setActionLoading(prev => ({ ...prev, [sessionId]: true }));
    
    try {
      await ApiService.completeCashierSession(sessionId);
      await loadActiveSessions(); // Перезагружаем список
    } catch (error) {
      console.error('Ошибка завершения сессии:', error);
      setError('Ошибка завершения сессии');
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  const handleCancelSession = async (sessionId) => {
    if (!window.confirm('Вы уверены, что хотите отменить эту сессию?')) {
      return;
    }

    setActionLoading(prev => ({ ...prev, [sessionId]: true }));
    
    try {
      await ApiService.cancelCashierSession(sessionId);
      await loadActiveSessions(); // Перезагружаем список
    } catch (error) {
      console.error('Ошибка отмены сессии:', error);
      setError('Ошибка отмены сессии');
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  const handleEnableChemistry = async (sessionId) => {
    if (!window.confirm('Вы уверены, что хотите включить химию для этой сессии?')) {
      return;
    }

    setActionLoading(prev => ({ ...prev, [sessionId]: true }));
    
    try {
      await ApiService.enableChemistryCashier(sessionId);
      await loadActiveSessions(); // Перезагружаем список
      alert('Химия успешно включена!');
    } catch (error) {
      console.error('Ошибка включения химии:', error);
      setError('Ошибка включения химии: ' + (error.response?.data?.error || error.message));
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  if (loading) {
    return <LoadingSpinner theme={theme}>Загрузка активных сессий...</LoadingSpinner>;
  }

  return (
    <Container theme={theme}>
      {error && (
        <ErrorMessage>{error}</ErrorMessage>
      )}

      <h3>Активные сессии</h3>
      
      {sessions.length === 0 ? (
        <EmptyMessage theme={theme}>
          Активных сессий нет
        </EmptyMessage>
      ) : (
        sessions.map(session => (
          <SessionCardComponent
            key={session.id}
            session={session}
            onStart={handleStartSession}
            onComplete={handleCompleteSession}
            onCancel={handleCancelSession}
            onEnableChemistry={handleEnableChemistry}
            actionLoading={actionLoading}
            theme={theme}
          />
        ))
      )}
    </Container>
  );
};

export default ActiveSessions; 