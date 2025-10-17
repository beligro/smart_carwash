import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useLocation, useNavigate } from 'react-router-dom';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

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
  
  &.expired {
    background-color: #fafafa;
    color: #616161;
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
    dateFrom: '',
    dateTo: '',
    userId: '',
    boxNumber: ''
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
      if (filters.dateFrom) filtersData.date_from = filters.dateFrom;
      if (filters.dateTo) filtersData.date_to = filters.dateTo;
      if (filters.userId) filtersData.user_id = filters.userId;
      if (filters.boxNumber) filtersData.box_number = filters.boxNumber;
      
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

  useEffect(() => {
    fetchSessions();
  }, [filters, pagination.limit, pagination.offset]);

  // Сброс пагинации при изменении фильтров
  useEffect(() => {
    setPagination(prev => ({ ...prev, offset: 0 }));
  }, [filters]);

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

  const handlePageChange = (newOffset) => {
    setPagination(prev => ({ ...prev, offset: newOffset }));
  };

  const getStatusText = (status) => {
    const statusMap = {
      created: 'Создана',
      assigned: 'Назначена',
      active: 'Активна',
      complete: 'Завершена',
      canceled: 'Отменена',
      expired: 'Истекла'
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
      air_dry: 'Обдув',
      vacuum: 'Пылесос'
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
      const date = new Date(dateString);
      if (isNaN(date.getTime())) {
        // Возвращаем исходную строку вместо "Некорректная дата"
        return dateString;
      }
      return date.toLocaleString('ru-RU');
    } catch (error) {
      console.error('Ошибка форматирования даты:', error, 'dateString:', dateString);
      // Возвращаем исходную строку вместо "Ошибка форматирования даты"
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

      <Filters>
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

        <FilterSelect
          value={filters.serviceType}
          onChange={(e) => setFilters({ ...filters, serviceType: e.target.value })}
        >
          <option value="">Все типы услуг</option>
          <option value="wash">Мойка</option>
          <option value="air_dry">Обдув</option>
          <option value="vacuum">Пылесос</option>
        </FilterSelect>

        <FilterInput
          type="text"
          value={filters.userId}
          onChange={(e) => setFilters({ ...filters, userId: e.target.value })}
          placeholder="ID клиента"
        />

        <FilterInput
          type="number"
          value={filters.boxNumber}
          onChange={(e) => setFilters({ ...filters, boxNumber: e.target.value })}
          placeholder="Номер бокса"
        />

        <FilterInput
          type="date"
          value={filters.dateFrom}
          onChange={(e) => setFilters({ ...filters, dateFrom: e.target.value })}
          placeholder="Дата от"
        />

        <FilterInput
          type="date"
          value={filters.dateTo}
          onChange={(e) => setFilters({ ...filters, dateTo: e.target.value })}
          placeholder="Дата до"
        />
      </Filters>

      <Table theme={theme}>
        <thead>
          <tr>
            <Th theme={theme}>ID</Th>
            <Th theme={theme}>Клиент</Th>
            <Th theme={theme}>Бокс</Th>
            <Th theme={theme}>Статус</Th>
            <Th theme={theme}>Тип услуги</Th>
            <Th theme={theme}>Химия</Th>
            <Th theme={theme}>Время аренды</Th>
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
              <Td>{session.rental_time_minutes} мин</Td>
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
              </Td>
            </tr>
          ))}
        </tbody>
      </Table>

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
                    <DetailLabel theme={theme}>Время аренды:</DetailLabel>
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
    </Container>
  );
};

export default SessionManagement; 