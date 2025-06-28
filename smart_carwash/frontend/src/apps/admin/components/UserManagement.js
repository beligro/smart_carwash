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

const AdminBadge = styled.span`
  background-color: #ff6b6b;
  color: white;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
`;

const UserManagement = () => {
  const theme = getTheme('light');
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [pagination, setPagination] = useState({
    limit: 20,
    offset: 0,
    total: 0
  });

  // Загрузка пользователей
  const fetchUsers = async () => {
    try {
      setLoading(true);
      setError('');
      
      const filters = {
        limit: pagination.limit,
        offset: pagination.offset
      };
      
      const response = await ApiService.getUsers(filters);
      setUsers(response.users || []);
      setPagination(prev => ({
        ...prev,
        total: response.total || 0
      }));
    } catch (err) {
      setError('Ошибка при загрузке пользователей');
      console.error('Error fetching users:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [pagination.limit, pagination.offset]);

  const handlePageChange = (newOffset) => {
    setPagination(prev => ({ ...prev, offset: newOffset }));
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString('ru-RU');
  };

  const totalPages = Math.ceil(pagination.total / pagination.limit);
  const currentPage = Math.floor(pagination.offset / pagination.limit) + 1;

  if (loading && users.length === 0) {
    return <LoadingMessage theme={theme}>Загрузка пользователей...</LoadingMessage>;
  }

  return (
    <Container>
      <Header>
        <Title theme={theme}>Управление пользователями</Title>
      </Header>

      {error && <ErrorMessage>{error}</ErrorMessage>}

      <Table theme={theme}>
        <thead>
          <tr>
            <Th theme={theme}>ID</Th>
            <Th theme={theme}>Telegram ID</Th>
            <Th theme={theme}>Имя пользователя</Th>
            <Th theme={theme}>Имя</Th>
            <Th theme={theme}>Фамилия</Th>
            <Th theme={theme}>Роль</Th>
            <Th theme={theme}>Дата регистрации</Th>
            <Th theme={theme}>Действия</Th>
          </tr>
        </thead>
        <tbody>
          {users.map((user) => (
            <tr key={user.id}>
              <Td>{user.id.substring(0, 8)}...</Td>
              <Td>{user.telegram_id}</Td>
              <Td>{user.username || '-'}</Td>
              <Td>{user.first_name || '-'}</Td>
              <Td>{user.last_name || '-'}</Td>
              <Td>
                {user.is_admin ? (
                  <AdminBadge>Администратор</AdminBadge>
                ) : (
                  'Пользователь'
                )}
              </Td>
              <Td>{formatDate(user.created_at)}</Td>
              <Td>
                <ActionButton theme={theme} onClick={() => window.open(`/admin/users/by-id?id=${user.id}`, '_blank')}>
                  Подробнее
                </ActionButton>
              </Td>
            </tr>
          ))}
        </tbody>
      </Table>

      {users.length === 0 && !loading && (
        <div style={{ textAlign: 'center', padding: '20px', color: theme.textColor }}>
          Пользователи не найдены
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
        Показано {users.length} из {pagination.total} пользователей
      </div>
    </Container>
  );
};

export default UserManagement; 