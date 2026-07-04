import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useNavigate, useLocation } from 'react-router-dom';
import { getTheme } from '../../shared/styles/theme';
import AuthService from '../../shared/services/AuthService';
import ApiService from '../../shared/services/ApiService';
import ActiveSessions from './components/ActiveSessions';
import LastShiftStatistics from './components/LastShiftStatistics';
import BoxManagement from './components/BoxManagement';
import QueueStatus from './components/QueueStatus';
import MobileTable from '../../shared/components/MobileTable';
import Timer from '../../shared/components/UI/Timer';
import useTimer from '../../shared/hooks/useTimer';
import usePolling from '../../shared/hooks/usePolling';
import ReassignSessionModal from '../../shared/components/UI/ReassignSessionModal/ReassignSessionModal';

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
  font-weight: 700;
  font-size: 1.15rem;
  color: ${props => props.theme?.primaryColor || '#007bff'};
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
  overflow-x: auto;
  
  @media (max-width: 768px) {
    flex-wrap: nowrap;
    gap: 0;
  }
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
  white-space: nowrap;
  min-width: fit-content;

  &:hover {
    color: ${props => props.theme.primaryColor};
  }
  
  @media (max-width: 768px) {
    padding: 12px 16px;
    font-size: 0.9rem;
    min-height: 44px;
  }
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
  
  @media (max-width: 768px) {
    display: none;
  }
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

const BoxNumberTd = styled.td`
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  font-size: 1.4rem;
  font-weight: 700;
`;

// Мобильные карточки для кассира
const MobileSessionCard = styled.div`
  display: none;
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border-left: 4px solid ${props => {
    switch (props.status) {
      case 'created': return '#ffc107';
      case 'in_queue': return '#17a2b8';
      case 'assigned': return '#007bff';
      case 'active': return '#28a745';
      case 'complete': return '#6c757d';
      case 'canceled': return '#dc3545';
      default: return '#6c757d';
    }
  }};
  
  @media (max-width: 768px) {
    display: block;
  }
`;

const MobileCardHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
`;

const MobileCardTitle = styled.div`
  font-weight: 600;
  font-size: 1rem;
  color: ${props => props.theme.textColor};
`;

const MobileCardStatus = styled.div`
  font-size: 0.8rem;
`;

const MobileCardDetails = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 12px;
`;

const MobileCardDetail = styled.div`
  display: flex;
  flex-direction: column;
`;

const MobileCardLabel = styled.span`
  font-size: 0.7rem;
  color: ${props => props.theme.textColorSecondary};
  margin-bottom: 2px;
`;

const MobileCardValue = styled.span`
  font-size: 0.8rem;
  color: ${props => props.theme.textColor};
  font-weight: 500;
`;

const MobileBoxNumberValue = styled.span`
  font-size: 1.4rem;
  color: ${props => props.theme.textColor};
  font-weight: 700;
`;

// Компонент для отображения таймера сессии
const SessionTimer = React.memo(({ session }) => {
  const { timeLeft } = useTimer(session);
  
  if (!timeLeft || timeLeft <= 0) {
    return null;
  }
  
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
      <span style={{ fontSize: '0.8rem', color: '#666' }}>⏱️</span>
      <Timer seconds={timeLeft} theme="light" />
    </div>
  );
});


const MobileCardActions = styled.div`
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
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
      default: return '#6c757d';
    }
  }};
  color: white;
`;

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

  &.complete {
    background-color: #dc3545;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #c82333;
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
  const [actionLoading, setActionLoading] = useState({});
  const [reassignModal, setReassignModal] = useState({
    isOpen: false,
    sessionId: null,
    serviceType: null
  });

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

  // Тихая перепроверка статуса смены без баннера ошибки. Нужна, чтобы при истечении
  // смены (09:00 НСК) кассира автоматически вернуло к экрану открытия смены,
  // даже если вкладка браузера оставалась открытой и не перезагружалась.
  const pollShiftStatus = async () => {
    try {
      const response = await ApiService.getShiftStatus();
      setHasActiveShift(response.has_active_shift);
      setShiftInfo(response.shift);
    } catch (error) {
      // Сетевые сбои поллинга игнорируем, чтобы не мешать работе кассира
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
    
    // Не загружаем данные для вкладок активных сессий, очереди и боксов - компоненты сами загружают
    if (activeTab === 'active_sessions' || activeTab === 'queue' || activeTab === 'boxes') {
      return;
    }
    
    setLoadingData(true);
    setError(null);
    
    try {
      if (activeTab === 'sessions') {
        console.log('Загружаем все сессии для кассира с начала смены:', shiftInfo.started_at);
        const sessionsResponse = await ApiService.getSessions({ 
          limit: 100,
          date_from: shiftInfo.started_at
        });
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

  const pollSessions = async () => {
    if (!shiftInfo || activeTab !== 'sessions') {
      return;
    }
    
    try {
      console.log('Поллинг сессий для кассира с начала смены:', shiftInfo.started_at);
      const sessionsResponse = await ApiService.getSessions({ 
        limit: 100,
        date_from: shiftInfo.started_at
      });
      const newSessions = sessionsResponse.sessions || [];
      
      // Обновляем только если данные изменились
      setSessions(prevSessions => {
        if (JSON.stringify(prevSessions) !== JSON.stringify(newSessions)) {
          return newSessions;
        }
        return prevSessions;
      });
    } catch (error) {
      console.error('Ошибка поллинга сессий:', error);
      // Не показываем ошибку при поллинге, чтобы не мешать пользователю
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

  const getStatusText = (status) => {
    switch (status) {
      case 'created': return 'Создана';
      case 'in_queue': return 'В очереди';
      case 'assigned': return 'Назначена';
      case 'active': return 'Активна';
      case 'complete': return 'Завершена';
      case 'canceled': return 'Отменена';
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

  const getChemistryStatus = (session) => {
    if (!session.with_chemistry) {
      return { text: 'Без химии', icon: '❌', color: '#6c757d' };
    }
    if (session.was_chemistry_on) {
      return { text: 'Химия включена', icon: '🧪✅', color: '#28a745' };
    }
    return { text: 'Химия оплачена', icon: '🧪', color: '#ffc107' };
  };

  // Обработчик завершения сессии из таблицы сессий
  const handleCompleteSessionFromTable = async (sessionId) => {
    if (!window.confirm('Вы уверены, что хотите завершить эту сессию?')) {
      return;
    }

    setActionLoading(prev => ({ ...prev, [sessionId]: true }));
    
    try {
      await ApiService.completeCashierSession(sessionId);
      await loadData(); // Перезагружаем данные
    } catch (error) {
      console.error('Ошибка завершения сессии:', error);
      setError('Ошибка завершения сессии');
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  // Обработчик отмены сессии из таблицы сессий
  const handleCancelSessionFromTable = async (sessionId) => {
    if (!window.confirm('Вы уверены, что хотите отменить эту сессию?')) {
      return;
    }

    setActionLoading(prev => ({ ...prev, [sessionId]: true }));
    
    try {
      await ApiService.cancelCashierSession(sessionId);
      await loadData(); // Перезагружаем данные
    } catch (error) {
      console.error('Ошибка отмены сессии:', error);
      setError('Ошибка отмены сессии: ' + (error.response?.data?.error || error.message));
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  // Обработчик переназначения сессии
  const handleReassignSession = async (sessionId) => {
    setActionLoading(prev => ({ ...prev, [sessionId]: true }));
    
    try {
      await ApiService.cashierReassignSession(sessionId);
      await loadData(); // Перезагружаем данные
      setReassignModal({ isOpen: false, sessionId: null, serviceType: null });
    } catch (error) {
      console.error('Ошибка переназначения сессии:', error);
      setError('Ошибка переназначения сессии: ' + (error.response?.data?.error || error.message));
    } finally {
      setActionLoading(prev => ({ ...prev, [sessionId]: false }));
    }
  };

  // Открытие модального окна переназначения
  const openReassignModal = (sessionId, serviceType) => {
    setReassignModal({
      isOpen: true,
      sessionId,
      serviceType
    });
  };

  // Закрытие модального окна переназначения
  const closeReassignModal = () => {
    setReassignModal({
      isOpen: false,
      sessionId: null,
      serviceType: null
    });
  };

  // Обработчик выхода из системы
  const handleLogout = async () => {
    await AuthService.logout();
    navigate('/cashier/login');
  };

  // Тихий поллинг для обновления данных без показа загрузки
  const pollData = async () => {
    console.log('pollData вызвана:', { shiftInfo, activeTab });
    if (!shiftInfo) {
      return;
    }
    
    // Не загружаем данные для вкладок активных сессий, очереди и боксов - компоненты сами загружают
    if (activeTab === 'active_sessions' || activeTab === 'queue' || activeTab === 'boxes') {
      return;
    }
    
    try {
      if (activeTab === 'sessions') {
        console.log('Поллинг сессий для кассира с начала смены:', shiftInfo.started_at);
        const sessionsResponse = await ApiService.getSessions({ 
          limit: 100,
          date_from: shiftInfo.started_at
        });
        const newSessions = sessionsResponse.sessions || [];
        
        // Обновляем только если данные изменились
        setSessions(prevSessions => {
          if (JSON.stringify(prevSessions) !== JSON.stringify(newSessions)) {
            return newSessions;
          }
          return prevSessions;
        });
      } else if (activeTab === 'payments') {
        console.log('Поллинг платежей для кассира, shiftStartedAt:', shiftInfo.started_at);
        const paymentsResponse = await ApiService.getCashierPayments(shiftInfo.started_at);
        const newPayments = paymentsResponse.payments || [];
        
        // Обновляем только если данные изменились
        setPayments(prevPayments => {
          if (JSON.stringify(prevPayments) !== JSON.stringify(newPayments)) {
            return newPayments;
          }
          return prevPayments;
        });
      }
    } catch (error) {
      console.error('Ошибка поллинга данных:', error);
      // Не показываем ошибку при поллинге, чтобы не мешать пользователю
    }
  };

  // Поллинг для автоматического обновления данных каждые 3 секунды (тихий)
  usePolling(pollData, 3000, hasActiveShift && !!shiftInfo, [activeTab]);

  // Поллинг для сессий каждые 3 секунды без показа загрузки (только для вкладки sessions)
  usePolling(pollSessions, 3000, hasActiveShift && !!shiftInfo && activeTab === 'sessions', [activeTab]);

  // Перепроверка статуса смены раз в минуту: при истечении смены в 09:00 НСК
  // hasActiveShift станет false и кассиру покажется экран открытия новой смены.
  usePolling(pollShiftStatus, 60000, hasActiveShift && !!shiftInfo, []);
  
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
              <Tab 
                active={activeTab === 'queue'} 
                onClick={() => handleTabChange('queue')}
                theme={theme}
              >
                Очередь
              </Tab>
              <Tab 
                active={activeTab === 'boxes'} 
                onClick={() => handleTabChange('boxes')}
                theme={theme}
              >
                Боксы
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
                    <>
                      {/* Десктопная таблица */}
                      <Table>
                        <thead>
                          <tr>
                            <Th theme={theme}>Статус</Th>
                            <Th theme={theme}>Тип услуги</Th>
                            <Th theme={theme}>Номер машины</Th>
                            <Th theme={theme}>Номер бокса</Th>
                            <Th theme={theme}>Химия</Th>
                            <Th theme={theme}>Время мойки</Th>
                            <Th theme={theme}>Таймер</Th>
                            <Th theme={theme}>Создана</Th>
                            <Th theme={theme}>Действия</Th>
                          </tr>
                        </thead>
                        <tbody>
                          {sessions.map(session => {
                            const chemistryStatus = getChemistryStatus(session);
                            return (
                              <tr key={session.id}>
                                <Td theme={theme}>
                                  <StatusBadge status={session.status}>
                                    {getStatusText(session.status)}
                                  </StatusBadge>
                                </Td>
                                <Td theme={theme}>{getServiceTypeText(session.service_type)}</Td>
                                <Td theme={theme}>{session.car_number || 'Не указан'}</Td>
                                <BoxNumberTd theme={theme}>{session.box_number ? `№${session.box_number}` : 'Не назначен'}</BoxNumberTd>
                                <Td theme={theme}>
                                  <ChemistryStatus theme={theme} style={{ color: chemistryStatus.color }}>
                                    <ChemistryIcon>{chemistryStatus.icon}</ChemistryIcon>
                                    {chemistryStatus.text}
                                  </ChemistryStatus>
                                </Td>
                                <Td theme={theme}>{session.rental_time_minutes} мин</Td>
                                <Td theme={theme}>
                                  <SessionTimer session={session} />
                                </Td>
                                <Td theme={theme}>{formatDateTime(session.created_at)}</Td>
                                <Td theme={theme}>
                                  <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
                                    {session.status === 'active' && (
                                      <ActionButton
                                        className="complete"
                                        onClick={() => handleCompleteSessionFromTable(session.id)}
                                        disabled={actionLoading[session.id]}
                                        style={{ padding: '4px 8px', fontSize: '1.1rem' }}
                                      >
                                        {actionLoading[session.id] ? 'Завершаем...' : 'Завершить'}
                                      </ActionButton>
                                    )}
                                    {(session.status === 'active') && (
                                      <ActionButton
                                        className="reassign"
                                        onClick={() => openReassignModal(session.id, session.service_type)}
                                        disabled={actionLoading[session.id]}
                                        style={{ padding: '4px 8px', fontSize: '0.8rem', backgroundColor: '#ff9800', color: 'white' }}
                                      >
                                        {actionLoading[session.id] ? 'Переназначаем...' : '🔄 Переназначить'}
                                      </ActionButton>
                                    )}
                                    {/* Кнопка отмены сессии (created, in_queue, assigned для кассирских; только created для telegram) */}
                                    {(session.status === 'created') && (
                                      <ActionButton
                                        className="cancel"
                                        onClick={() => handleCancelSessionFromTable(session.id)}
                                        disabled={actionLoading[session.id]}
                                        style={{ padding: '4px 8px', fontSize: '0.9rem', backgroundColor: '#dc3545', color: 'white' }}
                                      >
                                        {actionLoading[session.id] ? 'Отменяем...' : 'Отменить'}
                                      </ActionButton>
                                    )}
                                  </div>
                                </Td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </Table>

                      {/* Мобильные карточки */}
                      {sessions.map(session => {
                        const chemistryStatus = getChemistryStatus(session);
                        return (
                          <MobileSessionCard key={session.id} theme={theme} status={session.status}>
                            <MobileCardHeader>
                              <MobileCardTitle theme={theme}>
                                Сессия
                              </MobileCardTitle>
                              <MobileCardStatus>
                                <StatusBadge status={session.status}>
                                  {getStatusText(session.status)}
                                </StatusBadge>
                              </MobileCardStatus>
                            </MobileCardHeader>
                            
                            <MobileCardDetails>
                              <MobileCardDetail>
                                <MobileCardLabel theme={theme}>Тип услуги</MobileCardLabel>
                                <MobileCardValue theme={theme}>
                                  {getServiceTypeText(session.service_type)}
                                </MobileCardValue>
                              </MobileCardDetail>
                              
                              <MobileCardDetail>
                                <MobileCardLabel theme={theme}>Номер машины</MobileCardLabel>
                                <MobileCardValue theme={theme}>
                                  {session.car_number || 'Не указан'}
                                </MobileCardValue>
                              </MobileCardDetail>
                              
                              <MobileCardDetail>
                                <MobileCardLabel theme={theme}>Номер бокса</MobileCardLabel>
                                <MobileBoxNumberValue theme={theme}>
                                  {session.box_number ? `№${session.box_number}` : 'Не назначен'}
                                </MobileBoxNumberValue>
                              </MobileCardDetail>
                              
                              <MobileCardDetail>
                                <MobileCardLabel theme={theme}>Химия</MobileCardLabel>
                                <MobileCardValue theme={theme}>
                                  <ChemistryStatus theme={theme} style={{ color: chemistryStatus.color }}>
                                    <ChemistryIcon>{chemistryStatus.icon}</ChemistryIcon>
                                    {chemistryStatus.text}
                                  </ChemistryStatus>
                                </MobileCardValue>
                              </MobileCardDetail>
                              
                              <MobileCardDetail>
                                <MobileCardLabel theme={theme}>Время мойки</MobileCardLabel>
                                <MobileCardValue theme={theme}>
                                  {session.rental_time_minutes} мин
                                </MobileCardValue>
                              </MobileCardDetail>
                              
                              <MobileCardDetail>
                                <MobileCardLabel theme={theme}>Таймер</MobileCardLabel>
                                <MobileCardValue theme={theme}>
                                  <SessionTimer session={session} />
                                </MobileCardValue>
                              </MobileCardDetail>
                              
                              <MobileCardDetail>
                                <MobileCardLabel theme={theme}>Создана</MobileCardLabel>
                                <MobileCardValue theme={theme}>
                                  {formatDateTime(session.created_at)}
                                </MobileCardValue>
                              </MobileCardDetail>
                            </MobileCardDetails>
                            
                            <MobileCardActions>
                              {session.status === 'active' && (
                                <ActionButton
                                  className="complete"
                                  onClick={() => handleCompleteSessionFromTable(session.id)}
                                  disabled={actionLoading[session.id]}
                                  style={{ padding: '8px 16px', fontSize: '0.9rem', minHeight: '44px' }}
                                >
                                  {actionLoading[session.id] ? 'Завершаем...' : 'Завершить'}
                                </ActionButton>
                              )}
                              {(session.status === 'active') && (
                                <ActionButton
                                  className="reassign"
                                  onClick={() => openReassignModal(session.id, session.service_type)}
                                  disabled={actionLoading[session.id]}
                                  style={{ padding: '8px 16px', fontSize: '0.9rem', minHeight: '44px', backgroundColor: '#ff9800', color: 'white' }}
                                >
                                  {actionLoading[session.id] ? 'Переназначаем...' : '🔄 Переназначить'}
                                </ActionButton>
                              )}
                              {/* Кнопка отмены сессии (created, in_queue, assigned для кассирских; только created для telegram) */}
                              {(session.status === 'created') && (
                                <ActionButton
                                  className="cancel"
                                  onClick={() => handleCancelSessionFromTable(session.id)}
                                  disabled={actionLoading[session.id]}
                                  style={{ padding: '8px 16px', fontSize: '0.9rem', minHeight: '44px', backgroundColor: '#dc3545', color: 'white' }}
                                >
                                  {actionLoading[session.id] ? 'Отменяем...' : 'Отменить'}
                                </ActionButton>
                              )}
                            </MobileCardActions>
                          </MobileSessionCard>
                        );
                      })}
                    </>
                  )}
                </div>
              ) : activeTab === 'payments' ? (
                <div>
                  <h3>Платежи с начала смены</h3>
                  {payments.length === 0 ? (
                    <p>Платежей пока нет</p>
                  ) : (
                    <>
                      {/* Десктопная таблица */}
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

                      {/* Мобильные карточки */}
                      <MobileTable
                        data={payments}
                        columns={[
                          { key: 'id', label: 'ID', accessor: (item) => item.id },
                          { key: 'status', label: 'Статус', accessor: (item) => (
                            <StatusBadge status={item.status}>{item.status}</StatusBadge>
                          )},
                          { key: 'payment_type', label: 'Тип', accessor: (item) => item.payment_type },
                          { key: 'payment_method', label: 'Метод', accessor: (item) => item.payment_method },
                          { key: 'amount', label: 'Сумма', accessor: (item) => formatAmount(item.amount) },
                          { key: 'created_at', label: 'Создан', accessor: (item) => formatDateTime(item.created_at) }
                        ]}
                        getBorderColor={(payment) => {
                          switch (payment.status) {
                            case 'succeeded': return '#28a745';
                            case 'pending': return '#ffc107';
                            case 'failed': return '#dc3545';
                            default: return '#6c757d';
                          }
                        }}
                        theme={theme}
                        titleField="id"
                        statusField="status"
                      />
                    </>
                  )}
                </div>
              ) : activeTab === 'queue' ? (
                <QueueStatus />
              ) : activeTab === 'boxes' ? (
                <BoxManagement />
              ) : null}
            </Card>
          </>
        )}
      </Content>

      {/* Модальное окно переназначения сессии */}
      <ReassignSessionModal
        isOpen={reassignModal.isOpen}
        onClose={closeReassignModal}
        onConfirm={handleReassignSession}
        sessionId={reassignModal.sessionId}
        serviceType={reassignModal.serviceType}
        isLoading={actionLoading[reassignModal.sessionId] || false}
      />
    </CashierContainer>
  );
};

export default CashierApp;
