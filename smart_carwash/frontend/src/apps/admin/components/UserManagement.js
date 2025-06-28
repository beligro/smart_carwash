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

// Модальное окно для деталей пользователя
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

const UserDetails = styled.div`
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
  text-decoration: underline;
  font-size: 14px;
  padding: 0;
  margin: 0;
  
  &:hover {
    color: ${props => props.theme.primaryColorDark};
  }
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
  
  // Состояние для модального окна деталей пользователя
  const [showUserModal, setShowUserModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const [userDetails, setUserDetails] = useState(null);
  const [userDetailsLoading, setUserDetailsLoading] = useState(false);

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

  // Загрузка деталей пользователя
  const fetchUserDetails = async (userId) => {
    try {
      setUserDetailsLoading(true);
      // Здесь можно добавить API вызов для получения деталей пользователя
      // Пока используем данные из списка
      setUserDetails(selectedUser);
    } catch (err) {
      console.error('Error fetching user details:', err);
      setError('Ошибка при загрузке деталей пользователя');
    } finally {
      setUserDetailsLoading(false);
    }
  };

  // Открытие модального окна с деталями пользователя
  const openUserModal = async (user) => {
    setSelectedUser(user);
    setShowUserModal(true);
    await fetchUserDetails(user.id);
  };

  // Закрытие модального окна
  const closeUserModal = () => {
    setShowUserModal(false);
    setSelectedUser(null);
    setUserDetails(null);
    setError('');
  };

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
                <ActionButton theme={theme} onClick={() => openUserModal(user)}>
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

      {/* Модальное окно с деталями пользователя */}
      {showUserModal && (
        <Modal>
          <ModalContent>
            <ModalHeader>
              <ModalTitle theme={theme}>Детали пользователя</ModalTitle>
              <CloseButton onClick={closeUserModal} theme={theme}>
                &times;
              </CloseButton>
            </ModalHeader>
            
            {userDetailsLoading ? (
              <div style={{ textAlign: 'center', padding: '20px' }}>
                Загрузка деталей...
              </div>
            ) : userDetails ? (
              <div>
                <UserDetails>
                  <DetailGroup>
                    <DetailLabel theme={theme}>ID пользователя:</DetailLabel>
                    <DetailValue theme={theme}>{userDetails.id}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Telegram ID:</DetailLabel>
                    <DetailValue theme={theme}>{userDetails.telegram_id}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Имя пользователя:</DetailLabel>
                    <DetailValue theme={theme}>{userDetails.username || 'Не указано'}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Имя:</DetailLabel>
                    <DetailValue theme={theme}>{userDetails.first_name || 'Не указано'}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Фамилия:</DetailLabel>
                    <DetailValue theme={theme}>{userDetails.last_name || 'Не указано'}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Роль:</DetailLabel>
                    <DetailValue theme={theme}>
                      {userDetails.is_admin ? (
                        <AdminBadge>Администратор</AdminBadge>
                      ) : (
                        'Пользователь'
                      )}
                    </DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Дата регистрации:</DetailLabel>
                    <DetailValue theme={theme}>{formatDate(userDetails.created_at)}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Дата обновления:</DetailLabel>
                    <DetailValue theme={theme}>{formatDate(userDetails.updated_at)}</DetailValue>
                  </DetailGroup>
                </UserDetails>
                
                <div style={{ marginTop: '20px', padding: '15px', backgroundColor: '#f8f9fa', borderRadius: '5px' }}>
                  <h4 style={{ margin: '0 0 10px 0', color: theme.textColor }}>Полезные ссылки:</h4>
                  <div style={{ display: 'flex', gap: '15px', flexWrap: 'wrap' }}>
                    <LinkButton theme={theme} onClick={() => {
                      // Здесь можно добавить переход к сессиям пользователя
                      console.log('Переход к сессиям пользователя:', userDetails.id);
                    }}>
                      Посмотреть сессии пользователя
                    </LinkButton>
                  </div>
                </div>
              </div>
            ) : (
              <div style={{ textAlign: 'center', padding: '20px', color: theme.textColor }}>
                Детали пользователя не найдены
              </div>
            )}
          </ModalContent>
        </Modal>
      )}
    </Container>
  );
};

export default UserManagement; 