import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useNavigate } from 'react-router-dom';
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

const RefreshButton = styled.button`
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  
  &:hover {
    background-color: ${props => props.theme.primaryColorDark};
  }
  
  &:disabled {
    background-color: #ccc;
    cursor: not-allowed;
  }
`;

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
`;

const StatCard = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  text-align: center;
`;

const StatNumber = styled.div`
  font-size: 2rem;
  font-weight: bold;
  color: ${props => props.theme.primaryColor};
  margin-bottom: 5px;
`;

const StatLabel = styled.div`
  color: ${props => props.theme.textColor};
  font-size: 14px;
`;

const QueueSection = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
`;

const SectionTitle = styled.h3`
  margin: 0 0 15px 0;
  color: ${props => props.theme.textColor};
  display: flex;
  align-items: center;
  gap: 10px;
`;

const QueueBadge = styled.span`
  background-color: ${props => props.hasQueue ? '#ff6b6b' : '#51cf66'};
  color: white;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
`;

const BoxesGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 10px;
  margin-bottom: 15px;
`;

const BoxCard = styled.div`
  background-color: ${props => {
    switch (props.status) {
      case 'free': return '#e8f5e8';
      case 'busy': return '#ffe8e8';
      case 'reserved': return '#fff3e8';
      case 'maintenance': return '#e8f0ff';
      default: return '#f5f5f5';
    }
  }};
  border: 2px solid ${props => {
    switch (props.status) {
      case 'free': return '#2e7d32';
      case 'busy': return '#c62828';
      case 'reserved': return '#ef6c00';
      case 'maintenance': return '#1565c0';
      default: return '#ccc';
    }
  }};
  padding: 10px;
  border-radius: 8px;
  text-align: center;
  font-weight: 500;
`;

const BoxNumber = styled.div`
  font-size: 1.2rem;
  font-weight: bold;
  margin-bottom: 5px;
`;

const BoxStatus = styled.div`
  font-size: 12px;
  text-transform: uppercase;
`;

const UsersList = styled.div`
  margin-top: 15px;
`;

const UserItem = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background-color: #f8f9fa;
  border-radius: 5px;
  margin-bottom: 5px;
`;

const UserInfo = styled.div`
  display: flex;
  flex-direction: column;
`;

const UserName = styled.div`
  font-weight: 500;
  color: ${props => props.theme.textColor};
`;

const UserDetails = styled.div`
  font-size: 12px;
  color: #666;
`;

const UserPosition = styled.div`
  background-color: ${props => props.theme.primaryColor};
  color: white;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
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

const EmptyMessage = styled.div`
  color: ${props => props.theme.textColor};
  text-align: center;
  padding: 20px;
  font-size: 14px;
  font-style: italic;
`;

const QueueStatus = () => {
  const theme = getTheme('light');
  const [queueData, setQueueData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  // Загрузка данных очереди
  const fetchQueueStatus = async () => {
    try {
      setLoading(true);
      setError('');
      
      const response = await ApiService.getQueueStatus();
      setQueueData(response);
    } catch (err) {
      setError('Ошибка при загрузке статуса очереди');
      console.error('Error fetching queue status:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchQueueStatus();
  }, []);

  // Автообновление каждые 30 секунд
  useEffect(() => {
    const interval = setInterval(fetchQueueStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  const getStatusText = (status) => {
    const statusMap = {
      free: 'Свободен',
      busy: 'Занят',
      reserved: 'Зарезервирован',
      maintenance: 'Обслуживание'
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

  if (loading && !queueData) {
    return <LoadingMessage theme={theme}>Загрузка статуса очереди...</LoadingMessage>;
  }

  if (!queueData) {
    return <EmptyMessage theme={theme}>Данные недоступны</EmptyMessage>;
  }

  const { queue_status, details } = queueData;

  return (
    <Container>
      <Header>
        <Title theme={theme}>Текущее состояние очереди</Title>
        <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
          <RefreshButton theme={theme} onClick={fetchQueueStatus} disabled={loading}>
            {loading ? 'Обновление...' : 'Обновить'}
          </RefreshButton>
        </div>
      </Header>

      {error && <ErrorMessage>{error}</ErrorMessage>}

      <StatsGrid>
        <StatCard theme={theme}>
          <StatNumber theme={theme}>{queue_status.total_queue_size}</StatNumber>
          <StatLabel theme={theme}>Всего в очереди</StatLabel>
        </StatCard>
        
        <StatCard theme={theme}>
          <StatNumber theme={theme}>{queue_status.all_boxes.length}</StatNumber>
          <StatLabel theme={theme}>Всего боксов</StatLabel>
        </StatCard>
        
        <StatCard theme={theme}>
          <StatNumber theme={theme}>
            {queue_status.all_boxes.filter(box => box.status === 'free').length}
          </StatNumber>
          <StatLabel theme={theme}>Свободных боксов</StatLabel>
        </StatCard>
        
        <StatCard theme={theme}>
          <StatNumber theme={theme}>
            {queue_status.all_boxes.filter(box => box.status === 'busy').length}
          </StatNumber>
          <StatLabel theme={theme}>Занятых боксов</StatLabel>
        </StatCard>
      </StatsGrid>

      {/* Очередь мойки */}
      <QueueSection theme={theme}>
        <SectionTitle theme={theme}>
          Мойка
          <QueueBadge hasQueue={queue_status.wash_queue.has_queue}>
            {queue_status.wash_queue.has_queue ? 'Есть очередь' : 'Нет очереди'}
          </QueueBadge>
        </SectionTitle>
        
        <div style={{ marginBottom: '10px', color: theme.textColor }}>
          В очереди: {queue_status.wash_queue.queue_size} человек
        </div>
        
        <BoxesGrid>
          {queue_status.wash_queue.boxes.map((box) => (
            <BoxCard key={box.id} status={box.status}>
              <BoxNumber>№{box.number}</BoxNumber>
              <BoxStatus>{getStatusText(box.status)}</BoxStatus>
            </BoxCard>
          ))}
        </BoxesGrid>

        {queue_status.wash_queue.users_in_queue && queue_status.wash_queue.users_in_queue.length > 0 && (
          <UsersList>
            <h4 style={{ margin: '0 0 10px 0', color: theme.textColor }}>Пользователи в очереди:</h4>
            {queue_status.wash_queue.users_in_queue.map((user, index) => (
              <UserItem key={user.user_id}>
                <UserInfo theme={theme}>
                  <UserName theme={theme}>
                    {(() => {
                      let displayName = '';
                      if (user.first_name && user.last_name) {
                        displayName = `${user.first_name} ${user.last_name}`;
                      } else if (user.first_name) {
                        displayName = user.first_name;
                      } else if (user.last_name) {
                        displayName = user.last_name;
                      } else {
                        displayName = `Пользователь ${user.user_id.substring(0, 8)}`;
                      }
                      
                      const username = user.username ? ` (${user.username})` : '';
                      return displayName + username;
                    })()}
                  </UserName>
                  <UserDetails>
                    Ожидает с: {user.waiting_since}
                  </UserDetails>
                </UserInfo>
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                  <UserPosition theme={theme}>
                    Позиция {user.position}
                  </UserPosition>
                  <button
                    style={{
                      background: 'none',
                      border: 'none',
                      color: theme.primaryColor,
                      cursor: 'pointer',
                      fontSize: '12px',
                      textDecoration: 'underline'
                    }}
                    onClick={() => {
                      navigate(`/admin/users?highlight=${user.user_id}`);
                    }}
                  >
                    Подробнее
                  </button>
                </div>
              </UserItem>
            ))}
          </UsersList>
        )}
      </QueueSection>

      {/* Очередь обдува */}
      <QueueSection theme={theme}>
        <SectionTitle theme={theme}>
          Обдув
          <QueueBadge hasQueue={queue_status.air_dry_queue.has_queue}>
            {queue_status.air_dry_queue.has_queue ? 'Есть очередь' : 'Нет очереди'}
          </QueueBadge>
        </SectionTitle>
        
        <div style={{ marginBottom: '10px', color: theme.textColor }}>
          В очереди: {queue_status.air_dry_queue.queue_size} человек
        </div>
        
        <BoxesGrid>
          {queue_status.air_dry_queue.boxes.map((box) => (
            <BoxCard key={box.id} status={box.status}>
              <BoxNumber>№{box.number}</BoxNumber>
              <BoxStatus>{getStatusText(box.status)}</BoxStatus>
            </BoxCard>
          ))}
        </BoxesGrid>

        {queue_status.air_dry_queue.users_in_queue && queue_status.air_dry_queue.users_in_queue.length > 0 && (
          <UsersList>
            <h4 style={{ margin: '0 0 10px 0', color: theme.textColor }}>Пользователи в очереди:</h4>
            {queue_status.air_dry_queue.users_in_queue.map((user, index) => (
              <UserItem key={user.user_id}>
                <UserInfo theme={theme}>
                  <UserName theme={theme}>
                    {(() => {
                      let displayName = '';
                      if (user.first_name && user.last_name) {
                        displayName = `${user.first_name} ${user.last_name}`;
                      } else if (user.first_name) {
                        displayName = user.first_name;
                      } else if (user.last_name) {
                        displayName = user.last_name;
                      } else {
                        displayName = `Пользователь ${user.user_id.substring(0, 8)}`;
                      }
                      
                      const username = user.username ? ` (${user.username})` : '';
                      return displayName + username;
                    })()}
                  </UserName>
                  <UserDetails>
                    Ожидает с: {user.waiting_since}
                  </UserDetails>
                </UserInfo>
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                  <UserPosition theme={theme}>
                    Позиция {user.position}
                  </UserPosition>
                  <button
                    style={{
                      background: 'none',
                      border: 'none',
                      color: theme.primaryColor,
                      cursor: 'pointer',
                      fontSize: '12px',
                      textDecoration: 'underline'
                    }}
                    onClick={() => {
                      navigate(`/admin/users?highlight=${user.user_id}`);
                    }}
                  >
                    Подробнее
                  </button>
                </div>
              </UserItem>
            ))}
          </UsersList>
        )}
      </QueueSection>

      {/* Очередь пылесоса */}
      <QueueSection theme={theme}>
        <SectionTitle theme={theme}>
          Пылесос
          <QueueBadge hasQueue={queue_status.vacuum_queue.has_queue}>
            {queue_status.vacuum_queue.has_queue ? 'Есть очередь' : 'Нет очереди'}
          </QueueBadge>
        </SectionTitle>
        
        <div style={{ marginBottom: '10px', color: theme.textColor }}>
          В очереди: {queue_status.vacuum_queue.queue_size} человек
        </div>
        
        <BoxesGrid>
          {queue_status.vacuum_queue.boxes.map((box) => (
            <BoxCard key={box.id} status={box.status}>
              <BoxNumber>№{box.number}</BoxNumber>
              <BoxStatus>{getStatusText(box.status)}</BoxStatus>
            </BoxCard>
          ))}
        </BoxesGrid>

        {queue_status.vacuum_queue.users_in_queue && queue_status.vacuum_queue.users_in_queue.length > 0 && (
          <UsersList>
            <h4 style={{ margin: '0 0 10px 0', color: theme.textColor }}>Пользователи в очереди:</h4>
            {queue_status.vacuum_queue.users_in_queue.map((user, index) => (
              <UserItem key={user.user_id}>
                <UserInfo theme={theme}>
                  <UserName theme={theme}>
                    {(() => {
                      let displayName = '';
                      if (user.first_name && user.last_name) {
                        displayName = `${user.first_name} ${user.last_name}`;
                      } else if (user.first_name) {
                        displayName = user.first_name;
                      } else if (user.last_name) {
                        displayName = user.last_name;
                      } else {
                        displayName = `Пользователь ${user.user_id.substring(0, 8)}`;
                      }
                      
                      const username = user.username ? ` (${user.username})` : '';
                      return displayName + username;
                    })()}
                  </UserName>
                  <UserDetails>
                    Ожидает с: {user.waiting_since}
                  </UserDetails>
                </UserInfo>
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                  <UserPosition theme={theme}>
                    Позиция {user.position}
                  </UserPosition>
                  <button
                    style={{
                      background: 'none',
                      border: 'none',
                      color: theme.primaryColor,
                      cursor: 'pointer',
                      fontSize: '12px',
                      textDecoration: 'underline'
                    }}
                    onClick={() => {
                      navigate(`/admin/users?highlight=${user.user_id}`);
                    }}
                  >
                    Подробнее
                  </button>
                </div>
              </UserItem>
            ))}
          </UsersList>
        )}
      </QueueSection>

      <div style={{ marginTop: '20px', textAlign: 'center', color: theme.textColor, fontSize: '12px' }}>
        Данные обновляются автоматически каждые 30 секунд
      </div>
    </Container>
  );
};

export default QueueStatus; 