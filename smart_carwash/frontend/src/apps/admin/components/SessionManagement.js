import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
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

const SessionManagement = () => {
  const theme = getTheme('light');
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filters, setFilters] = useState({
    status: '',
    serviceType: '',
    dateFrom: '',
    dateTo: ''
  });
  const [pagination, setPagination] = useState({
    limit: 20,
    offset: 0,
    total: 0
  });

  // Загрузка сессий
  const fetchSessions = async () => {
    try {
      setLoading(true);
      setError('');
      
      const params = new URLSearchParams();
      if (filters.status) params.append('status', filters.status);
      if (filters.serviceType) params.append('service_type', filters.serviceType);
      if (filters.dateFrom) params.append('date_from', filters.dateFrom);
      if (filters.dateTo) params.append('date_to', filters.dateTo);
      params.append('limit', pagination.limit.toString());
      params.append('offset', pagination.offset.toString());
      
      const response = await ApiService.get(`/admin/sessions?${params.toString()}`);
      setSessions(response.data.sessions || []);
      setPagination(prev => ({
        ...prev,
        total: response.data.total || 0
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
    const serviceMap = {
      wash: 'Мойка',
      air_dry: 'Обдув',
      vacuum: 'Пылесос'
    };
    return serviceMap[serviceType] || serviceType;
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString('ru-RU');
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
            <Th theme={theme}>Пользователь</Th>
            <Th theme={theme}>Бокс</Th>
            <Th theme={theme}>Статус</Th>
            <Th theme={theme}>Тип услуги</Th>
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
              <Td>{session.rental_time_minutes} мин</Td>
              <Td>{formatDate(session.created_at)}</Td>
              <Td>
                <ActionButton theme={theme} onClick={() => window.open(`/admin/sessions/by-id?id=${session.id}`, '_blank')}>
                  Подробнее
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

      <div style={{ marginTop: '20px', textAlign: 'center', color: theme.textColor }}>
        Показано {sessions.length} из {pagination.total} сессий
      </div>
    </Container>
  );
};

export default SessionManagement; 