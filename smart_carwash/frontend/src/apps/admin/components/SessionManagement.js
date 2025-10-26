import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useLocation, useNavigate } from 'react-router-dom';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import { 
  convertDateTimeLocalToUTC, 
  convertUTCToDateTimeLocal, 
  getQuickFilterDates,
  formatDateForDisplay 
} from '../../../shared/utils/dateUtils';
import ReassignSessionModal from '../../../shared/components/UI/ReassignSessionModal/ReassignSessionModal';
import Timer from '../../../shared/components/UI/Timer';
import useTimer from '../../../shared/hooks/useTimer';

const Container = styled.div`
  padding: 20px;
`;

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
`;

const Title = styled.h2`
  margin: 0;
  color: ${props => props.theme.textColor};
`;

const Filters = styled.div`
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
`;

const FilterSelect = styled.select`
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 5px;
  background-color: white;
  font-size: 14px;
`;

const FilterInput = styled.input`
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 5px;
  background-color: white;
  font-size: 14px;
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

const Th = styled.th`
  background-color: ${props => props.theme.primaryColor};
  color: white;
  padding: 12px;
  text-align: left;
  font-weight: 500;
`;

const Td = styled.td`
  padding: 12px;
  border-bottom: 1px solid #eee;
`;

const StatusBadge = styled.span`
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
  
  &.created {
    background-color: #e3f2fd;
    color: #1565c0;
  }
  
  &.assigned {
    background-color: #fff3e0;
    color: #ef6c00;
  }
  
  &.active {
    background-color: #e8f5e8;
    color: #2e7d32;
  }
  
  &.complete {
    background-color: #f3e5f5;
    color: #7b1fa2;
  }
  
  &.canceled {
    background-color: #ffebee;
    color: #c62828;
  }
  
`;

const ServiceTypeBadge = styled.span`
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  
  &.wash {
    background-color: #e3f2fd;
    color: #1565c0;
  }
  
  &.air_dry {
    background-color: #f3e5f5;
    color: #7b1fa2;
  }
  
  &.vacuum {
    background-color: #e8f5e8;
    color: #2e7d32;
  }
`;

const ActionButton = styled.button`
  background: none;
  border: none;
  color: ${props => props.theme.primaryColor};
  cursor: pointer;
  margin-right: 10px;
  font-size: 14px;
  
  &:hover {
    text-decoration: underline;
  }
`;

const Button = styled.button`
  padding: 10px 20px;
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: background-color 0.2s ease;

  &:hover {
    background-color: ${props => props.theme.primaryColorHover};
  }

  &:disabled {
    background-color: ${props => props.theme.disabledColor};
    cursor: not-allowed;
  }
`;

const ErrorMessage = styled.div`
  color: #d32f2f;
  margin-bottom: 15px;
  font-size: 14px;
`;

const LoadingMessage = styled.div`
  color: ${props => props.theme.textColor};
  text-align: center;
  padding: 20px;
  font-size: 14px;
`;

const Pagination = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 20px;
`;

const PageButton = styled.button`
  background-color: ${props => props.active ? props.theme.primaryColor : 'white'};
  color: ${props => props.active ? 'white' : props.theme.textColor};
  border: 1px solid #ddd;
  padding: 8px 12px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  
  &:hover {
    background-color: ${props => props.active ? props.theme.primaryColorDark : '#f5f5f5'};
  }
  
  &:disabled {
    background-color: #f5f5f5;
    color: #ccc;
    cursor: not-allowed;
  }
`;

// Модальное окно для деталей сессии
const Modal = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
`;

const ModalContent = styled.div`
  background-color: white;
  padding: 30px;
  border-radius: 8px;
  width: 600px;
  max-width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  
  @media (max-width: 768px) {
    width: 95%;
    max-width: 95%;
    margin: 10px;
    padding: 20px;
    max-height: 90vh;
  }
`;

const ModalHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
`;

const ModalTitle = styled.h3`
  margin: 0;
  color: ${props => props.theme.textColor};
`;

const CloseButton = styled.button`
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: ${props => props.theme.textColor};
  
  &:hover {
    color: ${props => props.theme.primaryColor};
  }
`;

const SessionDetails = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
`;

const DetailGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: 5px;
`;

const DetailLabel = styled.span`
  font-weight: 500;
  color: ${props => props.theme.textColor};
  font-size: 14px;
`;

const DetailValue = styled.span`
  color: ${props => props.theme.textColor};
  font-size: 14px;
`;

const LinkButton = styled.button`
  background: none;
  border: none;
  color: ${props => props.theme.primaryColor};
  cursor: pointer;
  font-size: 14px;
  
  &:hover {
    text-decoration: underline;
  }
`;

// Мобильные стили для таблиц
const MobileCard = styled.div`
  display: none;
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border-left: 4px solid ${props => {
    switch (props.status) {
      case 'created': return '#ffc107';
      case 'assigned': return '#007bff';
      case 'active': return '#28a745';
      case 'complete': return '#6c757d';
      case 'canceled': return '#dc3545';
      case 'expired': return '#6c757d';
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
  color: ${props => props.theme.textColorSecondary};
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

const MobileCardActions = styled.div`
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
`;

const MobileActionButton = styled.button`
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  min-height: 32px;
  min-width: 32px;

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &.primary {
    background-color: ${props => props.theme.primaryColor};
    color: white;
    
    &:hover:not(:disabled) {
      opacity: 0.9;
    }
  }

  &.secondary {
    background-color: #6c757d;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #5a6268;
    }
  }
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

const SessionManagement = () => {
  const theme = getTheme('light');
  const location = useLocation();
  const navigate = useNavigate();
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filters, setFilters] = useState({
    status: '',
    serviceType: '',
    userId: '',
    boxNumber: ''
  });
  const [localDateFilters, setLocalDateFilters] = useState({
    dateFrom: '',
    dateTo: ''
  });
  const [pagination, setPagination] = useState({
    limit: 20,
    offset: 0,
    total: 0
  });
  
  // Состояние для модального окна деталей сессии
  const [showSessionModal, setShowSessionModal] = useState(false);
  const [selectedSession, setSelectedSession] = useState(null);
  const [sessionDetails, setSessionDetails] = useState(null);
  const [sessionDetailsLoading, setSessionDetailsLoading] = useState(false);

  // Состояние для модальных окон фильтров времени
  const [showDateFromModal, setShowDateFromModal] = useState(false);
  const [showDateToModal, setShowDateToModal] = useState(false);
  const [tempDateFrom, setTempDateFrom] = useState('');
  const [tempDateTo, setTempDateTo] = useState('');

  // Состояние для модального окна переназначения сессии
  const [reassignModal, setReassignModal] = useState({
    isOpen: false,
    sessionId: null,
    serviceType: null
  });
  const [reassignLoading, setReassignLoading] = useState(false);

  // Инициализация фильтров из state (при переходе от пользователей)
  useEffect(() => {
    if (location.state?.filters) {
      setFilters(prev => ({ ...prev, ...location.state.filters }));
    }
  }, [location.state]);

  // Загрузка сессий
  const fetchSessions = async () => {
    try {
      setLoading(true);
      setError('');
      
      // Исправляем названия полей для API
      const filtersData = {
        limit: pagination.limit,
        offset: pagination.offset
      };
      
      // Добавляем фильтры с правильными названиями полей
      if (filters.status) filtersData.status = filters.status;
      if (filters.serviceType) filtersData.service_type = filters.serviceType;
      if (filters.userId) filtersData.user_id = filters.userId;
      if (filters.boxNumber) filtersData.box_number = filters.boxNumber;
      
      // Конвертируем локальные даты в UTC перед отправкой
      if (localDateFilters.dateFrom) {
        filtersData.date_from = convertDateTimeLocalToUTC(localDateFilters.dateFrom);
      }
      if (localDateFilters.dateTo) {
        filtersData.date_to = convertDateTimeLocalToUTC(localDateFilters.dateTo);
      }
      
      const response = await ApiService.getSessions(filtersData);
      
      setSessions(response.sessions || []);
      setPagination(prev => ({
        ...prev,
        total: response.total || 0
      }));
    } catch (err) {
      setError('Ошибка при загрузке сессий');
      console.error('Error fetching sessions:', err);
    } finally {
      setLoading(false);
    }
  };

  // Поллинг сессий без показа загрузки
  const pollSessions = async () => {
    try {
      // Исправляем названия полей для API
      const filtersData = {
        limit: pagination.limit,
        offset: pagination.offset
      };
      
      // Добавляем фильтры с правильными названиями полей
      if (filters.status) filtersData.status = filters.status;
      if (filters.serviceType) filtersData.service_type = filters.serviceType;
      if (filters.userId) filtersData.user_id = filters.userId;
      if (filters.boxNumber) filtersData.box_number = filters.boxNumber;
      
      // Конвертируем локальные даты в UTC перед отправкой
      if (localDateFilters.dateFrom) {
        filtersData.date_from = convertDateTimeLocalToUTC(localDateFilters.dateFrom);
      }
      if (localDateFilters.dateTo) {
        filtersData.date_to = convertDateTimeLocalToUTC(localDateFilters.dateTo);
      }
      
      const response = await ApiService.getSessions(filtersData);
      
      setSessions(response.sessions || []);
      setPagination(prev => ({
        ...prev,
        total: response.total || 0
      }));
    } catch (err) {
      console.error('Ошибка поллинга сессий:', err);
      // Не показываем ошибку при поллинге, чтобы не мешать пользователю
    }
  };

  useEffect(() => {
    fetchSessions();
  }, [filters, localDateFilters, pagination.limit, pagination.offset]);

  // Поллинг для сессий каждые 3 секунды без показа загрузки
  useEffect(() => {
    const interval = setInterval(pollSessions, 3000);
    
    return () => clearInterval(interval);
  }, [filters, localDateFilters, pagination.limit, pagination.offset]);

  // Сброс пагинации при изменении фильтров
  useEffect(() => {
    setPagination(prev => ({ ...prev, offset: 0 }));
  }, [filters, localDateFilters]);

  // Загрузка деталей сессии
  const fetchSessionDetails = async (sessionId) => {
    try {
      setSessionDetailsLoading(true);
      const response = await ApiService.getSessionById(sessionId);
      setSessionDetails(response.session);
    } catch (err) {
      console.error('Error fetching session details:', err);
      setError('Ошибка при загрузке деталей сессии');
    } finally {
      setSessionDetailsLoading(false);
    }
  };

  // Открытие модального окна с деталями сессии
  const openSessionModal = async (session) => {
    setSelectedSession(session);
    setShowSessionModal(true);
    await fetchSessionDetails(session.id);
  };

  // Закрытие модального окна
  const closeSessionModal = () => {
    setShowSessionModal(false);
    setSelectedSession(null);
    setSessionDetails(null);
    setError('');
  };

  // Обработка поиска
  const handleSearch = () => {
    setPagination(prev => ({ ...prev, offset: 0 }));
    fetchSessions();
  };

  // Очистка фильтров
  const handleClearFilters = () => {
    setFilters({
      status: '',
      serviceType: '',
      userId: '',
      boxNumber: ''
    });
    setLocalDateFilters({
      dateFrom: '',
      dateTo: ''
    });
    setPagination(prev => ({ ...prev, offset: 0 }));
  };

  // Функции для модальных окон фильтров времени
  const openDateFromModal = () => {
    setTempDateFrom(localDateFilters.dateFrom);
    setShowDateFromModal(true);
  };

  const openDateToModal = () => {
    setTempDateTo(localDateFilters.dateTo);
    setShowDateToModal(true);
  };

  const applyDateFromFilter = () => {
    setLocalDateFilters(prev => ({ ...prev, dateFrom: tempDateFrom }));
    setShowDateFromModal(false);
    setPagination(prev => ({ ...prev, offset: 0 }));
  };

  const applyDateToFilter = () => {
    setLocalDateFilters(prev => ({ ...prev, dateTo: tempDateTo }));
    setShowDateToModal(false);
    setPagination(prev => ({ ...prev, offset: 0 }));
  };

  const cancelDateFromFilter = () => {
    setTempDateFrom(localDateFilters.dateFrom);
    setShowDateFromModal(false);
  };

  const cancelDateToFilter = () => {
    setTempDateTo(localDateFilters.dateTo);
    setShowDateToModal(false);
  };

  const handlePageChange = (newOffset) => {
    setPagination(prev => ({ ...prev, offset: newOffset }));
  };

  // Обработчики переназначения сессии
  const handleReassignSession = async (sessionId) => {
    setReassignLoading(true);
    
    try {
      await ApiService.adminReassignSession(sessionId);
      await fetchSessions(); // Перезагружаем список
      setReassignModal({ isOpen: false, sessionId: null, serviceType: null });
    } catch (error) {
      console.error('Ошибка переназначения сессии:', error);
      setError('Ошибка переназначения сессии: ' + (error.response?.data?.error || error.message));
    } finally {
      setReassignLoading(false);
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

  const getStatusText = (status) => {
    const statusMap = {
      created: 'Создана',
      assigned: 'Назначена',
      active: 'Активна',
      complete: 'Завершена',
      canceled: 'Отменена',
      in_queue: 'В очереди'
    };
    return statusMap[status] || status;
  };

  const getServiceTypeText = (serviceType) => {
    if (!serviceType) {
      console.warn('getServiceTypeText: serviceType is null or undefined');
      return 'Тип услуги не указан';
    }
    
    console.log('getServiceTypeText: serviceType =', serviceType);
    
    const serviceMap = {
      wash: 'Мойка',
      air_dry: 'Воздух для продувки',
      vacuum: 'Пылеводосос'
    };
    
    const result = serviceMap[serviceType];
    if (!result) {
      console.warn('getServiceTypeText: unknown serviceType =', serviceType);
      // Возвращаем исходное значение вместо "Неизвестная услуга"
      return serviceType;
    }
    
    return result;
  };

  const formatDate = (dateString) => {
    if (!dateString) {
      return 'Дата не указана';
    }
    
    try {
      return formatDateForDisplay(dateString);
    } catch (error) {
      console.error('Ошибка форматирования даты:', error, 'dateString:', dateString);
      return dateString;
    }
  };

  const totalPages = Math.ceil(pagination.total / pagination.limit);
  const currentPage = Math.floor(pagination.offset / pagination.limit) + 1;

  if (loading && sessions.length === 0) {
    return <LoadingMessage theme={theme}>Загрузка сессий...</LoadingMessage>;
  }

  return (
    <Container>
      <Header>
        <Title theme={theme}>Управление сессиями мойки</Title>
      </Header>

      {error && <ErrorMessage>{error}</ErrorMessage>}

      <Filters className="filters-container">
        <div className="filter-item">
          <FilterSelect
            value={filters.status}
            onChange={(e) => setFilters({ ...filters, status: e.target.value })}
          >
            <option value="">Все статусы</option>
            <option value="created">Создана</option>
            <option value="assigned">Назначена</option>
            <option value="active">Активна</option>
            <option value="complete">Завершена</option>
            <option value="canceled">Отменена</option>
            <option value="expired">Истекла</option>
          </FilterSelect>
        </div>

        <div className="filter-item">
          <FilterSelect
            value={filters.serviceType}
            onChange={(e) => setFilters({ ...filters, serviceType: e.target.value })}
          >
            <option value="">Все типы услуг</option>
            <option value="wash">Мойка</option>
            <option value="air_dry">Обдув</option>
            <option value="vacuum">Пылесос</option>
          </FilterSelect>
        </div>

        <div className="filter-item">
          <FilterInput
            type="text"
            value={filters.userId}
            onChange={(e) => setFilters({ ...filters, userId: e.target.value })}
            placeholder="ID клиента"
          />
        </div>

        <div className="filter-item">
          <FilterInput
            type="number"
            value={filters.boxNumber}
            onChange={(e) => setFilters({ ...filters, boxNumber: e.target.value })}
            placeholder="Номер бокса"
          />
        </div>

        <div className="filter-item">
          <Button 
            theme={theme} 
            onClick={openDateFromModal}
            style={{ 
              padding: '10px 20px', 
              backgroundColor: localDateFilters.dateFrom ? theme.primaryColor : '#f5f5f5',
              color: localDateFilters.dateFrom ? 'white' : theme.textColor,
              border: '1px solid #ddd',
              borderRadius: '4px',
              cursor: 'pointer',
              width: '100%',
              textAlign: 'left'
            }}
          >
            {localDateFilters.dateFrom ? `Дата от: ${localDateFilters.dateFrom}` : 'Выбрать дату от'}
          </Button>
        </div>

        <div className="filter-item">
          <Button 
            theme={theme} 
            onClick={openDateToModal}
            style={{ 
              padding: '10px 20px', 
              backgroundColor: localDateFilters.dateTo ? theme.primaryColor : '#f5f5f5',
              color: localDateFilters.dateTo ? 'white' : theme.textColor,
              border: '1px solid #ddd',
              borderRadius: '4px',
              cursor: 'pointer',
              width: '100%',
              textAlign: 'left'
            }}
          >
            {localDateFilters.dateTo ? `Дата до: ${localDateFilters.dateTo}` : 'Выбрать дату до'}
          </Button>
        </div>

        <div className="filter-item">
          <Button theme={theme} onClick={handleSearch}>
            Сохранить
          </Button>
        </div>

        <div className="filter-item">
          <Button theme={theme} onClick={handleClearFilters}>
            Очистить
          </Button>
        </div>
      </Filters>

      {/* Десктопная таблица */}
      <Table theme={theme} className="mobile-table">
        <thead>
          <tr>
            <Th theme={theme}>ID</Th>
            <Th theme={theme}>Клиент</Th>
            <Th theme={theme}>Бокс</Th>
            <Th theme={theme}>Статус</Th>
            <Th theme={theme}>Тип услуги</Th>
            <Th theme={theme}>Химия</Th>
            <Th theme={theme}>Время химии</Th>
            <Th theme={theme}>Время мойки</Th>
            <Th theme={theme}>Таймер</Th>
            <Th theme={theme}>Дата создания</Th>
            <Th theme={theme}>Действия</Th>
          </tr>
        </thead>
        <tbody>
          {sessions.map((session) => (
            <tr key={session.id}>
              <Td>{session.id.substring(0, 8)}...</Td>
              <Td>{session.user_id.substring(0, 8)}...</Td>
              <Td>{session.box_number || '-'}</Td>
              <Td>
                <StatusBadge className={session.status}>
                  {getStatusText(session.status)}
                </StatusBadge>
              </Td>
              <Td>
                <ServiceTypeBadge className={session.service_type}>
                  {getServiceTypeText(session.service_type)}
                </ServiceTypeBadge>
              </Td>
              <Td>
                {session.with_chemistry ? (
                  <span style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '4px',
                    fontSize: '12px'
                  }}>
                    <span style={{ color: '#4CAF50' }}>🧪</span>
                    {session.was_chemistry_on ? 'Включена' : 'Не включена'}
                  </span>
                ) : (
                  <span style={{ color: '#999', fontSize: '12px' }}>-</span>
                )}
              </Td>
              <Td>
                {session.with_chemistry ? (
                  <span style={{ fontSize: '12px' }}>{session.chemistry_time_minutes || 0} мин</span>
                ) : (
                  <span style={{ color: '#999', fontSize: '12px' }}>-</span>
                )}
              </Td>
              <Td>{session.rental_time_minutes} мин</Td>
              <Td>
                <SessionTimer session={session} />
              </Td>
              <Td>{formatDate(session.created_at)}</Td>
              <Td>
                <ActionButton theme={theme} onClick={() => openSessionModal(session)}>
                  Подробнее
                </ActionButton>
                <ActionButton 
                  theme={theme} 
                  onClick={() => navigate(`/admin/payments?session_id=${session.id}`)}
                  style={{ marginLeft: '8px', backgroundColor: '#10b981' }}
                >
                  Платежи
                </ActionButton>
                
                {/* Кнопка переназначения сессии */}
                {(session.status === 'active') && (
                  <ActionButton 
                    theme={theme} 
                    onClick={() => openReassignModal(session.id, session.service_type)}
                    style={{ marginLeft: '8px', backgroundColor: '#ff9800' }}
                    disabled={reassignLoading}
                  >
                    {reassignLoading ? 'Переназначаем...' : '🔄 Переназначить'}
                  </ActionButton>
                )}
              </Td>
            </tr>
          ))}
        </tbody>
      </Table>

      {/* Мобильные карточки */}
      <div className="mobile-card">
        {sessions.map((session) => (
          <MobileCard key={session.id} theme={theme} status={session.status}>
            <MobileCardHeader>
              <MobileCardTitle theme={theme}>
                Сессия #{session.id.substring(0, 8)}...
              </MobileCardTitle>
              <MobileCardStatus theme={theme}>
                <StatusBadge className={session.status}>
                  {getStatusText(session.status)}
                </StatusBadge>
              </MobileCardStatus>
            </MobileCardHeader>
            
            <MobileCardDetails>
              <MobileCardDetail>
                <MobileCardLabel theme={theme}>Клиент</MobileCardLabel>
                <MobileCardValue theme={theme}>
                  {session.user_id.substring(0, 8)}...
                </MobileCardValue>
              </MobileCardDetail>
              
              <MobileCardDetail>
                <MobileCardLabel theme={theme}>Бокс</MobileCardLabel>
                <MobileCardValue theme={theme}>
                  {session.box_number || '-'}
                </MobileCardValue>
              </MobileCardDetail>
              
              <MobileCardDetail>
                <MobileCardLabel theme={theme}>Тип услуги</MobileCardLabel>
                <MobileCardValue theme={theme}>
                  <ServiceTypeBadge className={session.service_type}>
                    {getServiceTypeText(session.service_type)}
                  </ServiceTypeBadge>
                </MobileCardValue>
              </MobileCardDetail>
              
              <MobileCardDetail>
                <MobileCardLabel theme={theme}>Химия</MobileCardLabel>
                <MobileCardValue theme={theme}>
                  {session.with_chemistry ? (
                    <span style={{ 
                      display: 'flex', 
                      alignItems: 'center', 
                      gap: '4px',
                      fontSize: '12px'
                    }}>
                      <span style={{ color: '#4CAF50' }}>🧪</span>
                      {session.was_chemistry_on ? 'Включена' : 'Не включена'}
                    </span>
                  ) : (
                    <span style={{ color: '#999', fontSize: '12px' }}>-</span>
                  )}
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
                  {formatDate(session.created_at)}
                </MobileCardValue>
              </MobileCardDetail>
            </MobileCardDetails>
            
            <MobileCardActions>
              <MobileActionButton 
                className="primary"
                onClick={() => openSessionModal(session)}
              >
                Подробнее
              </MobileActionButton>
              <MobileActionButton 
                className="secondary"
                onClick={() => navigate(`/admin/payments?session_id=${session.id}`)}
              >
                Платежи
              </MobileActionButton>
            </MobileCardActions>
          </MobileCard>
        ))}
      </div>

      {sessions.length === 0 && !loading && (
        <div style={{ textAlign: 'center', padding: '20px', color: theme.textColor }}>
          Сессии не найдены
        </div>
      )}

      {totalPages > 1 && (
        <Pagination>
          <PageButton
            theme={theme}
            disabled={currentPage === 1}
            onClick={() => handlePageChange(pagination.offset - pagination.limit)}
          >
            Предыдущая
          </PageButton>
          
          {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
            const page = Math.max(1, Math.min(totalPages - 4, currentPage - 2)) + i;
            return (
              <PageButton
                key={page}
                theme={theme}
                active={page === currentPage}
                onClick={() => handlePageChange((page - 1) * pagination.limit)}
              >
                {page}
              </PageButton>
            );
          })}
          
          <PageButton
            theme={theme}
            disabled={currentPage === totalPages}
            onClick={() => handlePageChange(pagination.offset + pagination.limit)}
          >
            Следующая
          </PageButton>
        </Pagination>
      )}

      {/* Модальное окно с деталями сессии */}
      {showSessionModal && (
        <Modal>
          <ModalContent>
            <ModalHeader>
              <ModalTitle theme={theme}>Детали сессии</ModalTitle>
              <CloseButton onClick={closeSessionModal} theme={theme}>
                &times;
              </CloseButton>
            </ModalHeader>
            
            {sessionDetailsLoading ? (
              <div style={{ textAlign: 'center', padding: '20px' }}>
                Загрузка деталей...
              </div>
            ) : sessionDetails ? (
              <div>
                <SessionDetails>
                  <DetailGroup>
                    <DetailLabel theme={theme}>ID сессии:</DetailLabel>
                    <DetailValue theme={theme}>{sessionDetails.id}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>ID клиента:</DetailLabel>
                    <DetailValue theme={theme}>{sessionDetails.user_id}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Номер бокса:</DetailLabel>
                    <DetailValue theme={theme}>{sessionDetails.box_number || 'Не назначен'}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Статус:</DetailLabel>
                    <DetailValue theme={theme}>
                      <StatusBadge className={sessionDetails.status}>
                        {getStatusText(sessionDetails.status)}
                      </StatusBadge>
                    </DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Тип услуги:</DetailLabel>
                    <DetailValue theme={theme}>
                      <ServiceTypeBadge className={sessionDetails.service_type}>
                        {getServiceTypeText(sessionDetails.service_type)}
                        {sessionDetails.with_chemistry && ' с химией'}
                      </ServiceTypeBadge>
                    </DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Номер машины:</DetailLabel>
                    <DetailValue theme={theme}>{sessionDetails.car_number || 'Не указан'}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Время мойки:</DetailLabel>
                    <DetailValue theme={theme}>{sessionDetails.rental_time_minutes} минут</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>С химией:</DetailLabel>
                    <DetailValue theme={theme}>
                      {sessionDetails.with_chemistry ? 'Да' : 'Нет'}
                    </DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Дата создания:</DetailLabel>
                    <DetailValue theme={theme}>{formatDate(sessionDetails.created_at)}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Дата обновления:</DetailLabel>
                    <DetailValue theme={theme}>{formatDate(sessionDetails.updated_at)}</DetailValue>
                  </DetailGroup>
                  
                  {sessionDetails.started_at && (
                    <DetailGroup>
                      <DetailLabel theme={theme}>Дата начала:</DetailLabel>
                      <DetailValue theme={theme}>{formatDate(sessionDetails.started_at)}</DetailValue>
                    </DetailGroup>
                  )}
                  
                  {sessionDetails.completed_at && (
                    <DetailGroup>
                      <DetailLabel theme={theme}>Дата завершения:</DetailLabel>
                      <DetailValue theme={theme}>{formatDate(sessionDetails.completed_at)}</DetailValue>
                    </DetailGroup>
                  )}
                  
                  {sessionDetails.extension_time_minutes && (
                    <DetailGroup>
                      <DetailLabel theme={theme}>Продление времени:</DetailLabel>
                      <DetailValue theme={theme}>{sessionDetails.extension_time_minutes} минут</DetailValue>
                    </DetailGroup>
                  )}
                </SessionDetails>
                
                <div style={{ marginTop: '20px', padding: '15px', backgroundColor: '#f8f9fa', borderRadius: '5px' }}>
                  <h4 style={{ margin: '0 0 10px 0', color: theme.textColor }}>Полезные ссылки:</h4>
                  <div style={{ display: 'flex', gap: '15px', flexWrap: 'wrap' }}>
                    <LinkButton theme={theme} onClick={() => {
                      closeSessionModal();
                      navigate(`/admin/users?highlight=${sessionDetails.user_id}`);
                    }}>
                      Посмотреть клиента
                    </LinkButton>
                    {sessionDetails.box_number && (
                      <LinkButton theme={theme} onClick={() => {
                        closeSessionModal();
                        navigate(`/admin/washboxes?highlight=${sessionDetails.box_number}`);
                      }}>
                        Посмотреть бокс №{sessionDetails.box_number}
                      </LinkButton>
                    )}
                    <LinkButton theme={theme} onClick={() => {
                      closeSessionModal();
                      navigate('/admin/sessions', { 
                        state: { 
                          filters: { userId: sessionDetails.user_id },
                          showUserFilter: true 
                        } 
                      });
                    }}>
                      Все сессии клиента
                    </LinkButton>
                  </div>
                </div>
              </div>
            ) : (
              <div style={{ textAlign: 'center', padding: '20px', color: theme.textColor }}>
                Детали сессии не найдены
              </div>
            )}
          </ModalContent>
        </Modal>
      )}

      {/* Модальное окно для выбора даты "от" */}
      {showDateFromModal && (
        <Modal>
          <ModalContent>
            <ModalHeader>
              <ModalTitle theme={theme}>Выберите дату "от"</ModalTitle>
              <CloseButton onClick={cancelDateFromFilter} theme={theme}>×</CloseButton>
            </ModalHeader>
            
            <div style={{ marginBottom: '20px' }}>
              <label style={{ display: 'block', marginBottom: '10px', fontWeight: '500' }}>
                Дата и время:
              </label>
              <input
                type="datetime-local"
                value={tempDateFrom}
                onChange={(e) => setTempDateFrom(e.target.value)}
                style={{
                  width: '100%',
                  padding: '10px',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                  fontSize: '16px'
                }}
              />
            </div>

            <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
              <Button theme={theme} onClick={cancelDateFromFilter} style={{ backgroundColor: '#6c757d' }}>
                Отмена
              </Button>
              <Button theme={theme} onClick={applyDateFromFilter}>
                Применить
              </Button>
            </div>
          </ModalContent>
        </Modal>
      )}

      {/* Модальное окно для выбора даты "до" */}
      {showDateToModal && (
        <Modal>
          <ModalContent>
            <ModalHeader>
              <ModalTitle theme={theme}>Выберите дату "до"</ModalTitle>
              <CloseButton onClick={cancelDateToFilter} theme={theme}>×</CloseButton>
            </ModalHeader>
            
            <div style={{ marginBottom: '20px' }}>
              <label style={{ display: 'block', marginBottom: '10px', fontWeight: '500' }}>
                Дата и время:
              </label>
              <input
                type="datetime-local"
                value={tempDateTo}
                onChange={(e) => setTempDateTo(e.target.value)}
                style={{
                  width: '100%',
                  padding: '10px',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                  fontSize: '16px'
                }}
              />
            </div>

            <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
              <Button theme={theme} onClick={cancelDateToFilter} style={{ backgroundColor: '#6c757d' }}>
                Отмена
              </Button>
              <Button theme={theme} onClick={applyDateToFilter}>
                Применить
              </Button>
            </div>
          </ModalContent>
        </Modal>
      )}

      <ReassignSessionModal
        isOpen={reassignModal.isOpen}
        onClose={closeReassignModal}
        onConfirm={handleReassignSession}
        sessionId={reassignModal.sessionId}
        serviceType={reassignModal.serviceType}
        isLoading={reassignLoading}
      />
    </Container>
  );
};

export default SessionManagement; 