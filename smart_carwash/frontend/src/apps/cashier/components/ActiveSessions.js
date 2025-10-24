import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import Timer from '../../../shared/components/UI/Timer';
import useTimer from '../../../shared/hooks/useTimer';
import ReassignSessionModal from '../../../shared/components/UI/ReassignSessionModal/ReassignSessionModal';

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

// Компонент кнопки включения химии для кассира (только кнопка!)
const ChemistryEnableButton = ({ session, onEnable, actionLoading }) => {
  return (
    <ActionButton
      className="chemistry"
      onClick={() => onEnable(session.id)}
      disabled={actionLoading[session.id]}
      style={{ backgroundColor: '#4CAF50', color: 'white' }}
    >
      {actionLoading[session.id] ? 'Включение...' : `🧪 Включить химию (${session.chemistry_time_minutes} мин)`}
    </ActionButton>
  );
};

// Компонент для отображения таймера химии в кассирском интерфейсе
const ChemistryTimer = ({ session }) => {
  const [chemistryTimeLeft, setChemistryTimeLeft] = useState(null);

  useEffect(() => {
    if (!session || !session.chemistry_started_at || session.chemistry_ended_at) {
      setChemistryTimeLeft(null);
      return;
    }

    const updateChemistryTimer = () => {
      const startTime = new Date(session.chemistry_started_at);
      const now = new Date();
      const timeLimit = (session.chemistry_time_minutes || 0) * 60 * 1000;
      const timePassed = now - startTime;
      const remaining = timeLimit - timePassed;

      if (remaining <= 0) {
        setChemistryTimeLeft(0);
      } else {
        setChemistryTimeLeft(Math.floor(remaining / 1000));
      }
    };

    updateChemistryTimer();
    const interval = setInterval(updateChemistryTimer, 1000);

    return () => clearInterval(interval);
  }, [session]);

  // Если химия выключена - ничего не показываем (статус уже в ChemistryStatus выше)
  if (session.chemistry_ended_at) {
    return null;
  }

  // Если химия активна - показываем таймер
  if (session.was_chemistry_on && session.chemistry_started_at && chemistryTimeLeft !== null && chemistryTimeLeft > 0) {
    return (
      <div style={{ marginTop: '8px', padding: '8px', backgroundColor: '#e8f5e9', borderRadius: '6px', border: '1px solid #4caf50' }}>
        <div style={{ fontSize: '11px', color: '#2e7d32', fontWeight: 'bold' }}>
          🧪 Химия: {Math.floor(chemistryTimeLeft / 60)}:{(chemistryTimeLeft % 60).toString().padStart(2, '0')}
        </div>
      </div>
    );
  }

  return null;
};

const ChemistryStatus = styled.div`
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.9rem;
  color: ${props => props.theme.textColor};
  font-weight: 500;
`;

const ChemistryIcon = styled.span`
  font-size: 1rem;
`;

/**
 * Компонент для отображения одной сессии
 */
const SessionCardComponent = ({ session, onStart, onComplete, onCancel, onEnableChemistry, onReassign, actionLoading, theme }) => {
  const { timeLeft } = useTimer(session);

  const getChemistryStatus = () => {
    if (!session.with_chemistry) {
      return { text: 'Без химии', icon: '❌', color: '#6c757d' };
    }
    const chemistryTime = session.chemistry_time_minutes || 0;
    if (session.was_chemistry_on) {
      return { text: `Химия включена (${chemistryTime} мин)`, icon: '🧪✅', color: '#28a745' };
    }
    return { text: `Химия оплачена (${chemistryTime} мин)`, icon: '🧪', color: '#ffc107' };
  };

  const chemistryStatus = getChemistryStatus();

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
        
        {/* Таймер химии (если активна) */}
        {session.status === 'active' && session.with_chemistry && session.was_chemistry_on && (
          <ChemistryTimer session={session} />
        )}
      </SessionHeader>

      <SessionDetails>
        <DetailItem>
          <DetailLabel theme={theme}>Тип услуги</DetailLabel>
          <DetailValue theme={theme}>
            {getServiceTypeText(session.service_type)}
          </DetailValue>
        </DetailItem>

        <DetailItem>
          <DetailLabel theme={theme}>Номер машины</DetailLabel>
          <DetailValue theme={theme}>
            {session.car_number || 'Не указан'}
          </DetailValue>
        </DetailItem>

        <DetailItem>
          <DetailLabel theme={theme}>Номер бокса</DetailLabel>
          <DetailValue theme={theme}>
            {session.box_number ? `Бокс ${session.box_number}` : 'Не назначен'}
          </DetailValue>
        </DetailItem>

        <DetailItem>
          <DetailLabel theme={theme}>Химия</DetailLabel>
          <ChemistryStatus theme={theme} style={{ color: chemistryStatus.color }}>
            <ChemistryIcon>{chemistryStatus.icon}</ChemistryIcon>
            {chemistryStatus.text}
          </ChemistryStatus>
        </DetailItem>

        <DetailItem>
          <DetailLabel theme={theme}>Время мойки</DetailLabel>
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

        {/* Кнопка переназначения сессии */}
        {(session.status === 'assigned' || session.status === 'active') && (
          <ActionButton
            className="reassign"
            onClick={() => onReassign(session.id, session.service_type)}
            disabled={actionLoading[session.id]}
            style={{ backgroundColor: '#ff9800', color: 'white' }}
          >
            {actionLoading[session.id] ? 'Переназначаем...' : '🔄 Переназначить'}
          </ActionButton>
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
  const [reassignModal, setReassignModal] = useState({
    isOpen: false,
    sessionId: null,
    serviceType: null
  });

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
    } catch (error) {
      console.error('Ошибка включения химии:', error);
      setError('Ошибка включения химии: ' + (error.response?.data?.error || error.message));
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  const handleReassignSession = async (sessionId) => {
    setActionLoading(prev => ({ ...prev, [sessionId]: true }));
    
    try {
      await ApiService.cashierReassignSession(sessionId);
      await loadActiveSessions(); // Перезагружаем список
      setReassignModal({ isOpen: false, sessionId: null, serviceType: null });
    } catch (error) {
      console.error('Ошибка переназначения сессии:', error);
      setError('Ошибка переназначения сессии: ' + (error.response?.data?.error || error.message));
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  const openReassignModal = (sessionId, serviceType) => {
    setReassignModal({
      isOpen: true,
      sessionId,
      serviceType
    });
  };

  const closeReassignModal = () => {
    setReassignModal({
      isOpen: false,
      sessionId: null,
      serviceType: null
    });
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
            onReassign={openReassignModal}
            actionLoading={actionLoading}
            theme={theme}
          />
        ))
      )}

      <ReassignSessionModal
        isOpen={reassignModal.isOpen}
        onClose={closeReassignModal}
        onConfirm={handleReassignSession}
        sessionId={reassignModal.sessionId}
        serviceType={reassignModal.serviceType}
        isLoading={actionLoading[reassignModal.sessionId] || false}
      />
    </Container>
  );
};

export default ActiveSessions; 