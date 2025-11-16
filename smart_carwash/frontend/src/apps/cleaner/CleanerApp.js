import React, { useState, useEffect, useCallback, useRef } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../shared/styles/theme';
import AuthService from '../../shared/services/AuthService';
import ApiService from '../../shared/services/ApiService';

const CleanerContainer = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: ${props => props.theme.backgroundColor};
  color: ${props => props.theme.textColor};
`;

const Header = styled.header`
  background-color: ${props => props.theme.cardBackground};
  padding: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const Title = styled.h1`
  margin: 0;
  font-size: 1.5rem;
`;

const LogoutButton = styled.button`
  background: none;
  border: none;
  color: ${props => props.theme.textColor};
  cursor: pointer;
  font-size: 1rem;
  
  &:hover {
    color: ${props => props.theme.primaryColor};
  }
`;

const UserInfo = styled.div`
  display: flex;
  align-items: center;
`;

const Username = styled.span`
  margin-right: 15px;
  font-weight: 500;
`;

const Content = styled.main`
  flex: 1;
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  
  @media (max-width: 768px) {
    padding: 16px;
  }
`;

const Card = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  
  @media (max-width: 768px) {
    padding: 16px;
    margin-bottom: 16px;
  }
`;

const ErrorMessage = styled.div`
  color: #dc3545;
  padding: 16px;
  background-color: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 6px;
  margin-bottom: 16px;
`;

const StatusBadge = styled.span`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 110px;
  padding: 6px 12px;
  border-radius: 999px;
  font-weight: 600;
  font-size: 0.9rem;
  color: #fff;
  background-color: ${props => {
    switch (props.status) {
      case 'cleaning':
        return '#17a2b8';
      case 'maintenance':
        return '#6c757d';
      case 'free':
        return '#28a745';
      default:
        return '#6c757d';
    }
  }};
`;

const InfoGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
  margin-top: 16px;
`;

const InfoBlock = styled.div`
  padding: 16px;
  border-radius: 8px;
  background-color: ${props => props.theme.backgroundColor};
  border: 1px solid ${props => props.theme.borderColor};
`;

const InfoLabel = styled.div`
  color: ${props => props.theme.textColorSecondary};
  font-size: 0.85rem;
  margin-bottom: 4px;
`;

const InfoValue = styled.div`
  font-size: 1.1rem;
  font-weight: 600;
`;

const ActionsRow = styled.div`
  margin-top: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
`;

const PrimaryButton = styled.button`
  padding: 12px 20px;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  color: white;
  background-color: ${props => (props.variant === 'stop' ? '#dc3545' : '#28a745')};
  opacity: ${props => (props.disabled ? 0.6 : 1)};
  pointer-events: ${props => (props.disabled ? 'none' : 'auto')};

  &:hover {
    filter: brightness(0.95);
  }
`;

const SecondaryButton = styled.button`
  padding: 12px 20px;
  border-radius: 6px;
  border: 1px solid ${props => props.theme.borderColor};
  background-color: rgba(0, 0, 0, 0.05);
  color: ${props => props.theme.textColor};
  cursor: pointer;
  font-weight: 500;
  opacity: ${props => (props.disabled ? 0.6 : 1)};
  transition: background-color 0.2s ease;

  &:hover {
    background-color: rgba(0, 0, 0, 0.1);
  }
`;

const HistoryHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
`;

const HistoryTable = styled.table`
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
`;

const HistoryTh = styled.th`
  text-align: left;
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  font-weight: 600;
  color: ${props => props.theme.textColor};
`;

const HistoryTd = styled.td`
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  color: ${props => props.theme.textColor};
`;

const EmptyState = styled.div`
  padding: 40px 0;
  text-align: center;
  color: ${props => props.theme.textColorSecondary};
  font-style: italic;
`;

const Loader = styled.div`
  padding: 24px 0;
  text-align: center;
  color: ${props => props.theme.textColorSecondary};
`;

const HistoryTableWrapper = styled.div`
  margin-top: 16px;

  @media (max-width: 768px) {
    display: none;
  }
`;

const HistoryMobileList = styled.div`
  display: none;
  margin-top: 16px;

  @media (max-width: 768px) {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
`;

const HistoryMobileCard = styled.div`
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 8px;
  padding: 16px;
  background-color: ${props => props.theme.backgroundColor};
`;

const HistoryMobileRow = styled.div`
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;

  &:last-child {
    margin-bottom: 0;
  }
`;

const HistoryMobileLabel = styled.span`
  color: ${props => props.theme.textColorSecondary};
  font-size: 0.9rem;
`;

const HistoryMobileValue = styled.span`
  font-weight: 600;
`;

/**
 * Приложение уборщика
 * @returns {React.ReactNode} - Приложение уборщика
 */
const HISTORY_LIMIT = 20;
const DEFAULT_CLEANING_TIMEOUT = 3;

const CleanerApp = () => {
  const theme = getTheme('light');
  const [user, setUser] = useState(null);
  const [authLoading, setAuthLoading] = useState(true);
  const [globalError, setGlobalError] = useState(null);
  const [boxState, setBoxState] = useState(null);
  const [boxLoading, setBoxLoading] = useState(true);
  const [actionInProgress, setActionInProgress] = useState(false);
  const [history, setHistory] = useState([]);
  const [historyMeta, setHistoryMeta] = useState({ total: 0, limit: HISTORY_LIMIT, offset: 0 });
  const [historyLoading, setHistoryLoading] = useState(true);
  const [historyError, setHistoryError] = useState(null);
  const [cleaningTimeout, setCleaningTimeout] = useState(DEFAULT_CLEANING_TIMEOUT);
  const [timerTick, setTimerTick] = useState(Date.now());
  const prevStatusRef = useRef(null);

  const getErrorMessage = (error, defaultMessage) => {
    if (error?.response?.data?.error) {
      return error.response.data.error;
    }
    return defaultMessage;
  };

  const loadHistory = useCallback(async () => {
    setHistoryLoading(true);
    try {
      setHistoryError(null);
      const response = await ApiService.getCleanerCleaningHistory({
        limit: HISTORY_LIMIT,
        offset: 0
      });
      setHistory(response?.logs || []);
      setHistoryMeta({
        total: response?.total || 0,
        limit: response?.limit || HISTORY_LIMIT,
        offset: response?.offset || 0
      });
    } catch (error) {
      setHistoryError(getErrorMessage(error, 'Не удалось загрузить историю уборок'));
    } finally {
      setHistoryLoading(false);
    }
  }, []);

  const loadBoxState = useCallback(async (showLoader = true) => {
    if (showLoader) {
      setBoxLoading(true);
    }
    try {
      setGlobalError(null);
      const response = await ApiService.getCleanerBoxState();
      const newState = response?.wash_box || null;
      const prevStatus = prevStatusRef.current;

      setBoxState(newState);

      if (prevStatus === 'cleaning' && newState?.status === 'maintenance') {
        loadHistory();
      }

      prevStatusRef.current = newState?.status || null;
    } catch (error) {
      setGlobalError(getErrorMessage(error, 'Не удалось загрузить состояние бокса'));
    } finally {
      setBoxLoading(false);
    }
  }, [loadHistory]);

  const loadCleaningTimeout = useCallback(async () => {
    try {
      const response = await ApiService.getCleaningTimeout();
      setCleaningTimeout(response?.timeout_minutes || DEFAULT_CLEANING_TIMEOUT);
    } catch (error) {
      console.warn('Не удалось загрузить время уборки, используем значение по умолчанию', error);
      setCleaningTimeout(DEFAULT_CLEANING_TIMEOUT);
    }
  }, []);

  useEffect(() => {
    const checkAuth = () => {
      const currentUser = AuthService.getCurrentUser();
      const isAuthenticated = AuthService.isAuthenticated();
      const isAdmin = AuthService.isAdmin();

      if (!isAuthenticated) {
        window.location.href = '/cleaner/login';
        return;
      }

      if (isAdmin) {
        window.location.href = '/admin';
        return;
      }

      setUser(currentUser);
      setAuthLoading(false);
    };

    checkAuth();
  }, []);

  useEffect(() => {
    if (authLoading) {
      return;
    }

    loadBoxState();
    loadHistory();
    loadCleaningTimeout();

    const interval = setInterval(() => {
      loadBoxState(false);
    }, 5000);

    return () => clearInterval(interval);
  }, [authLoading, loadBoxState, loadHistory, loadCleaningTimeout]);

  useEffect(() => {
    const interval = setInterval(() => setTimerTick(Date.now()), 1000);
    return () => clearInterval(interval);
  }, []);

  const handleLogout = async () => {
    await AuthService.logout();
    window.location.href = '/cleaner/login';
  };

  const handleStartCleaning = async () => {
    if (!boxState?.id || actionInProgress) {
      return;
    }

    setActionInProgress(true);
    try {
      await ApiService.startCleaning(boxState.id);
      await loadBoxState();
      await loadHistory();
    } catch (error) {
      setGlobalError(getErrorMessage(error, 'Не удалось начать уборку'));
    } finally {
      setActionInProgress(false);
    }
  };

  const handleCompleteCleaning = async () => {
    if (!boxState?.id || actionInProgress) {
      return;
    }

    setActionInProgress(true);
    try {
      await ApiService.completeCleaning(boxState.id);
      await loadBoxState();
      await loadHistory();
    } catch (error) {
      setGlobalError(getErrorMessage(error, 'Не удалось завершить уборку'));
    } finally {
      setActionInProgress(false);
    }
  };

  if (authLoading) {
    return null;
  }

  const formatDateTime = (value) => {
    if (!value) {
      return '—';
    }
    return new Date(value).toLocaleString('ru-RU');
  };

  const isCleaning = boxState?.status === 'cleaning';
  const formatTimer = () => {
    if (!isCleaning || !boxState?.cleaning_started_at) {
      return '—';
    }

    const startedAt = new Date(boxState.cleaning_started_at).getTime();
    const totalMs = cleaningTimeout * 60 * 1000;
    const elapsed = timerTick - startedAt;
    const remaining = Math.max(0, totalMs - elapsed);

    const minutes = Math.floor(remaining / 60000)
      .toString()
      .padStart(2, '0');
    const seconds = Math.floor((remaining % 60000) / 1000)
      .toString()
      .padStart(2, '0');

    return `${minutes}:${seconds}`;
  };
  
  return (
    <CleanerContainer theme={theme}>
      <Header theme={theme}>
        <Title>Интерфейс уборщика</Title>
        {user && (
          <UserInfo>
            <Username>{user.username}</Username>
            <LogoutButton onClick={handleLogout} theme={theme}>
              Выйти
            </LogoutButton>
          </UserInfo>
        )}
      </Header>
      <Content>
        {globalError && (
          <ErrorMessage>{globalError}</ErrorMessage>
        )}

        <Card theme={theme}>
          <h2>Бокс 40</h2>
          {boxLoading ? (
            <Loader theme={theme}>Загружаем состояние...</Loader>
          ) : boxState ? (
            <>
              <StatusBadge status={boxState.status}>
                {boxState.status === 'cleaning' && 'На уборке'}
                {boxState.status === 'maintenance' && 'На обслуживании'}
                {boxState.status === 'free' && 'Свободен'}
                {boxState.status !== 'cleaning' && boxState.status !== 'maintenance' && boxState.status !== 'free' && boxState.status}
              </StatusBadge>

              <InfoGrid>
                <InfoBlock theme={theme}>
                  <InfoLabel theme={theme}>Номер бокса</InfoLabel>
                  <InfoValue>№{boxState.number}</InfoValue>
                </InfoBlock>
                <InfoBlock theme={theme}>
                  <InfoLabel theme={theme}>Тип услуги</InfoLabel>
                  <InfoValue>{boxState.service_type === 'wash' ? 'Мойка' : boxState.service_type}</InfoValue>
                </InfoBlock>
                <InfoBlock theme={theme}>
                  <InfoLabel theme={theme}>Время начала уборки</InfoLabel>
                  <InfoValue>{formatDateTime(boxState.cleaning_started_at)}</InfoValue>
                </InfoBlock>
                <InfoBlock theme={theme}>
                  <InfoLabel theme={theme}>Оставшееся время</InfoLabel>
                  <InfoValue>{formatTimer()}</InfoValue>
                </InfoBlock>
              </InfoGrid>

              <ActionsRow>
                {isCleaning ? (
                  <PrimaryButton
                    variant="stop"
                    onClick={handleCompleteCleaning}
                    disabled={actionInProgress}
                  >
                    {actionInProgress ? 'Завершаем...' : 'Закончить уборку'}
                  </PrimaryButton>
                ) : (
                  <PrimaryButton
                    onClick={handleStartCleaning}
                    disabled={actionInProgress}
                  >
                    {actionInProgress ? 'Запускаем...' : 'Начать уборку'}
                  </PrimaryButton>
                )}
                <SecondaryButton
                  onClick={() => loadBoxState()}
                  disabled={actionInProgress}
                  theme={theme}
                >
                  Обновить состояние
                </SecondaryButton>
              </ActionsRow>
            </>
          ) : (
            <EmptyState theme={theme}>
              Бокс недоступен. Обратитесь к администратору.
            </EmptyState>
          )}
        </Card>

        <Card theme={theme}>
          <HistoryHeader>
            <h2>История уборок</h2>
            <div>
              <span style={{ color: theme.textColorSecondary, fontSize: '0.9rem', marginRight: '12px' }}>
                Всего записей: {historyMeta.total}
              </span>
              <SecondaryButton onClick={loadHistory} disabled={historyLoading} theme={theme}>
                Обновить
              </SecondaryButton>
            </div>
          </HistoryHeader>

          {historyError && (
            <ErrorMessage>{historyError}</ErrorMessage>
          )}

          {historyLoading ? (
            <Loader theme={theme}>Загружаем историю...</Loader>
          ) : history.length === 0 ? (
            <EmptyState theme={theme}>
              История уборок пока пуста
            </EmptyState>
          ) : (
            <>
              <HistoryTableWrapper>
                <HistoryTable>
                  <thead>
                    <tr>
                      <HistoryTh theme={theme}>Начало</HistoryTh>
                      <HistoryTh theme={theme}>Окончание</HistoryTh>
                    </tr>
                  </thead>
                  <tbody>
                    {history.map((log) => (
                      <tr key={log.id}>
                        <HistoryTd theme={theme}>{formatDateTime(log.started_at)}</HistoryTd>
                        <HistoryTd theme={theme}>{formatDateTime(log.completed_at)}</HistoryTd>
                      </tr>
                    ))}
                  </tbody>
                </HistoryTable>
              </HistoryTableWrapper>

              <HistoryMobileList>
                {history.map((log) => (
                  <HistoryMobileCard key={log.id} theme={theme}>
                    <HistoryMobileRow>
                      <HistoryMobileLabel theme={theme}>Начало</HistoryMobileLabel>
                      <HistoryMobileValue>{formatDateTime(log.started_at)}</HistoryMobileValue>
                    </HistoryMobileRow>
                    <HistoryMobileRow>
                      <HistoryMobileLabel theme={theme}>Окончание</HistoryMobileLabel>
                      <HistoryMobileValue>{formatDateTime(log.completed_at)}</HistoryMobileValue>
                    </HistoryMobileRow>
                  </HistoryMobileCard>
                ))}
              </HistoryMobileList>
            </>
          )}
        </Card>
      </Content>
    </CleanerContainer>
  );
};

export default CleanerApp;

