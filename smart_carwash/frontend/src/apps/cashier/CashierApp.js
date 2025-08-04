import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useNavigate, useLocation } from 'react-router-dom';
import { getTheme } from '../../shared/styles/theme';
import AuthService from '../../shared/services/AuthService';
import ApiService from '../../shared/services/ApiService';
import ActiveSessions from './components/ActiveSessions';
import LastShiftStatistics from './components/LastShiftStatistics';

const CashierContainer = styled.div`
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

const Content = styled.main`
  flex: 1;
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
`;

const Card = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
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

const ShiftButton = styled.button`
  background-color: ${props => props.hasActiveShift ? '#dc3545' : '#28a745'};
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;

  &:hover {
    background-color: ${props => props.hasActiveShift ? '#c82333' : '#218838'};
  }

  &:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }
`;

const ShiftInfo = styled.div`
  margin-top: 16px;
  padding: 16px;
  background-color: ${props => props.theme.backgroundColor};
  border-radius: 6px;
  border-left: 4px solid #28a745;
`;

const ShiftStatus = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
`;

const ShiftTime = styled.div`
  font-size: 0.9rem;
  color: ${props => props.theme.textColorSecondary};
`;

const TabContainer = styled.div`
  display: flex;
  margin-bottom: 20px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
`;

const Tab = styled.button`
  background: none;
  border: none;
  padding: 12px 24px;
  cursor: pointer;
  font-size: 1rem;
  color: ${props => props.active ? props.theme.primaryColor : props.theme.textColor};
  border-bottom: 2px solid ${props => props.active ? props.theme.primaryColor : 'transparent'};
  transition: all 0.2s;

  &:hover {
    color: ${props => props.theme.primaryColor};
  }
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
`;

const Th = styled.th`
  text-align: left;
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  font-weight: 600;
`;

const Td = styled.td`
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
`;

const StatusBadge = styled.span`
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  background-color: ${props => {
    switch (props.status) {
      case 'created': return '#ffc107';
      case 'in_queue': return '#17a2b8';
      case 'assigned': return '#007bff';
      case 'active': return '#28a745';
      case 'complete': return '#6c757d';
      case 'canceled': return '#dc3545';
      case 'expired': return '#6c757d';
      default: return '#6c757d';
    }
  }};
  color: white;
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

/**
 * Приложение кассира
 * @returns {React.ReactNode} - Приложение кассира
 */
const CashierApp = () => {
  const theme = getTheme('light');
  const navigate = useNavigate();
  const location = useLocation();
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [hasActiveShift, setHasActiveShift] = useState(false);
  const [shiftInfo, setShiftInfo] = useState(null);
  const [activeTab, setActiveTab] = useState('active_sessions');
  const [sessions, setSessions] = useState([]);
  const [payments, setPayments] = useState([]);
  const [loadingShift, setLoadingShift] = useState(false);
  const [loadingData, setLoadingData] = useState(false);
  const [error, setError] = useState(null);
  const [showLastShiftStatistics, setShowLastShiftStatistics] = useState(false);

  useEffect(() => {
    // Проверяем авторизацию при загрузке компонента
    const checkAuth = () => {
      const currentUser = AuthService.getCurrentUser();
      const isAuthenticated = AuthService.isAuthenticated();
      const isAdmin = AuthService.isAdmin();
      
      // Если пользователь не авторизован, перенаправляем на страницу входа
      if (!isAuthenticated) {
        navigate('/cashier/login', { replace: true });
        return;
      }
      
      // Если пользователь администратор, перенаправляем на страницу администратора
      if (isAdmin) {
        navigate('/admin', { replace: true });
        return;
      }
      
      // Если все проверки пройдены, устанавливаем пользователя
      setUser(currentUser);
      setIsLoading(false);
    };
    
    checkAuth();
  }, [navigate]);

  useEffect(() => {
    if (user) {
      checkShiftStatus();
    }
  }, [user]);

  useEffect(() => {
    console.log('useEffect для loadData:', { hasActiveShift, shiftInfo, activeTab });
    if (hasActiveShift && shiftInfo) {
      loadData();
    }
  }, [hasActiveShift, shiftInfo, activeTab]);

  const checkShiftStatus = async () => {
    try {
      const response = await ApiService.getShiftStatus();
      setHasActiveShift(response.has_active_shift);
      setShiftInfo(response.shift);
    } catch (error) {
      console.error('Ошибка проверки статуса смены:', error);
      setError('Ошибка проверки статуса смены');
    }
  };

  const handleStartShift = async () => {
    setLoadingShift(true);
    setError(null);
    
    try {
      await ApiService.startShift();
      await checkShiftStatus();
    } catch (error) {
      console.error('Ошибка начала смены:', error);
      setError(error.response?.data?.error || 'Ошибка начала смены');
    } finally {
      setLoadingShift(false);
    }
  };

  const handleEndShift = async () => {
    setLoadingShift(true);
    setError(null);
    
    try {
      await ApiService.endShift();
      setHasActiveShift(false);
      setShiftInfo(null);
      setSessions([]);
      setPayments([]);
    } catch (error) {
      console.error('Ошибка завершения смены:', error);
      setError(error.response?.data?.error || 'Ошибка завершения смены');
    } finally {
      setLoadingShift(false);
    }
  };

  const loadData = async () => {
    console.log('loadData вызвана:', { shiftInfo, activeTab });
    if (!shiftInfo) {
      console.log('shiftInfo отсутствует, выход из loadData');
      return;
    }
    
    // Не загружаем данные для вкладки активных сессий - компонент сам загружает
    if (activeTab === 'active_sessions') {
      return;
    }
    
    setLoadingData(true);
    setError(null);
    
    try {
      if (activeTab === 'sessions') {
        console.log('Загружаем сессии для кассира, shiftStartedAt:', shiftInfo.started_at);
        const sessionsResponse = await ApiService.getCashierSessions(shiftInfo.started_at);
        console.log('Получены сессии:', sessionsResponse);
        setSessions(sessionsResponse.sessions || []);
      } else if (activeTab === 'payments') {
        console.log('Загружаем платежи для кассира, shiftStartedAt:', shiftInfo.started_at);
        const paymentsResponse = await ApiService.getCashierPayments(shiftInfo.started_at);
        console.log('Получены платежи:', paymentsResponse);
        setPayments(paymentsResponse.payments || []);
      }
    } catch (error) {
      console.error('Ошибка загрузки данных:', error);
      setError('Ошибка загрузки данных');
    } finally {
      setLoadingData(false);
    }
  };

  const handleTabChange = (tab) => {
    console.log('Смена вкладки на:', tab);
    setActiveTab(tab);
  };

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

  // Обработчик выхода из системы
  const handleLogout = async () => {
    await AuthService.logout();
    navigate('/cashier/login');
  };
  
  // Если идет загрузка, показываем пустой контент
  if (isLoading) {
    return null;
  }
  
  return (
    <CashierContainer theme={theme}>
      <Header theme={theme}>
        <Title>Интерфейс кассира</Title>
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
        {error && (
          <ErrorMessage>{error}</ErrorMessage>
        )}

        <Card theme={theme}>
          <h2>Управление сменой</h2>
          
          {!hasActiveShift ? (
            <div>
              <p>У вас нет активной смены. Нажмите кнопку ниже, чтобы начать смену.</p>
              <ShiftButton 
                onClick={handleStartShift} 
                disabled={loadingShift}
                hasActiveShift={false}
              >
                {loadingShift ? 'Начинаем смену...' : 'Начать смену'}
              </ShiftButton>
              
              {/* Кнопка для показа статистики последней смены */}
              <div style={{ marginTop: '20px' }}>
                <button
                  onClick={() => setShowLastShiftStatistics(!showLastShiftStatistics)}
                  style={{
                    background: 'none',
                    border: '1px solid #007bff',
                    color: '#007bff',
                    padding: '8px 16px',
                    borderRadius: '4px',
                    cursor: 'pointer',
                    fontSize: '0.9rem'
                  }}
                >
                  {showLastShiftStatistics ? 'Скрыть статистику' : 'Показать статистику последней смены'}
                </button>
                
                {showLastShiftStatistics && (
                  <LastShiftStatistics onClose={() => setShowLastShiftStatistics(false)} />
                )}
              </div>
            </div>
          ) : (
            <div>
              <ShiftStatus>
                <div>
                  <strong>Смена активна</strong>
                </div>
                <ShiftButton 
                  onClick={handleEndShift} 
                  disabled={loadingShift}
                  hasActiveShift={true}
                >
                  {loadingShift ? 'Завершаем смену...' : 'Завершить смену'}
                </ShiftButton>
              </ShiftStatus>
              
              {shiftInfo && (
                <ShiftInfo theme={theme}>
                  <ShiftTime theme={theme}>
                    <strong>Начало смены:</strong> {formatDateTime(shiftInfo.started_at)}
                  </ShiftTime>
                  <ShiftTime theme={theme}>
                    <strong>Истекает:</strong> {formatDateTime(shiftInfo.expires_at)}
                  </ShiftTime>
                </ShiftInfo>
              )}
            </div>
          )}
        </Card>

        {hasActiveShift && (
          <>
            <TabContainer theme={theme}>
              <Tab 
                active={activeTab === 'active_sessions'} 
                onClick={() => handleTabChange('active_sessions')}
                theme={theme}
              >
                Активные сессии
              </Tab>
              <Tab 
                active={activeTab === 'sessions'} 
                onClick={() => handleTabChange('sessions')}
                theme={theme}
              >
                Сессии
              </Tab>
              <Tab 
                active={activeTab === 'payments'} 
                onClick={() => handleTabChange('payments')}
                theme={theme}
              >
                Платежи
              </Tab>
            </TabContainer>

            <Card theme={theme}>
              {loadingData ? (
                <LoadingSpinner theme={theme}>Загрузка данных...</LoadingSpinner>
              ) : activeTab === 'active_sessions' ? (
                <ActiveSessions />
              ) : activeTab === 'sessions' ? (
                <div>
                  <h3>Сессии с начала смены</h3>
                  {sessions.length === 0 ? (
                    <p>Сессий пока нет</p>
                  ) : (
                    <Table>
                      <thead>
                        <tr>
                          <Th theme={theme}>ID</Th>
                          <Th theme={theme}>Статус</Th>
                          <Th theme={theme}>Тип услуги</Th>
                          <Th theme={theme}>Номер машины</Th>
                          <Th theme={theme}>Время аренды</Th>
                          <Th theme={theme}>Создана</Th>
                        </tr>
                      </thead>
                      <tbody>
                        {sessions.map(session => (
                          <tr key={session.id}>
                            <Td theme={theme}>{session.id}</Td>
                            <Td theme={theme}>
                              <StatusBadge status={session.status}>
                                {session.status}
                              </StatusBadge>
                            </Td>
                            <Td theme={theme}>{session.service_type}</Td>
                            <Td theme={theme}>{session.car_number}</Td>
                            <Td theme={theme}>{session.rental_time_minutes} мин</Td>
                            <Td theme={theme}>{formatDateTime(session.created_at)}</Td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  )}
                </div>
              ) : activeTab === 'payments' ? (
                <div>
                  <h3>Платежи с начала смены</h3>
                  {payments.length === 0 ? (
                    <p>Платежей пока нет</p>
                  ) : (
                    <Table>
                      <thead>
                        <tr>
                          <Th theme={theme}>ID</Th>
                          <Th theme={theme}>Статус</Th>
                          <Th theme={theme}>Тип</Th>
                          <Th theme={theme}>Метод</Th>
                          <Th theme={theme}>Сумма</Th>
                          <Th theme={theme}>Создан</Th>
                        </tr>
                      </thead>
                      <tbody>
                        {payments.map(payment => (
                          <tr key={payment.id}>
                            <Td theme={theme}>{payment.id}</Td>
                            <Td theme={theme}>
                              <StatusBadge status={payment.status}>
                                {payment.status}
                              </StatusBadge>
                            </Td>
                            <Td theme={theme}>{payment.payment_type}</Td>
                            <Td theme={theme}>{payment.payment_method}</Td>
                            <Td theme={theme}>{formatAmount(payment.amount)}</Td>
                            <Td theme={theme}>{formatDateTime(payment.created_at)}</Td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  )}
                </div>
              ) : null}
            </Card>
          </>
        )}
      </Content>
    </CashierContainer>
  );
};

export default CashierApp;
