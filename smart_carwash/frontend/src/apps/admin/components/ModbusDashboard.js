import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import axios from 'axios';

// Создаем простой API клиент для Modbus
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL ? process.env.REACT_APP_API_URL.replace('/v1', '') : '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Добавляем токен авторизации
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

const Container = styled.div`
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
`;

const Title = styled.h1`
  color: ${props => props.theme.textPrimary};
  margin-bottom: 24px;
  font-size: 24px;
  font-weight: 600;
`;

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
`;

const StatCard = styled.div`
  background: ${props => props.theme.backgroundSecondary};
  border: 1px solid ${props => props.theme.border};
  border-radius: 8px;
  padding: 20px;
`;

const StatTitle = styled.h3`
  color: ${props => props.theme.textSecondary};
  font-size: 14px;
  font-weight: 500;
  margin: 0 0 8px 0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
`;

const StatValue = styled.div`
  color: ${props => props.theme.textPrimary};
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 4px;
`;

const StatSubtext = styled.div`
  color: ${props => props.theme.textSecondary};
  font-size: 12px;
`;

const StatusIndicator = styled.div`
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: ${props => props.connected ? '#4CAF50' : '#F44336'};
  margin-right: 8px;
`;

const Section = styled.div`
  background: ${props => props.theme.backgroundSecondary};
  border: 1px solid ${props => props.theme.border};
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
`;

const SectionTitle = styled.h2`
  color: ${props => props.theme.textPrimary};
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 16px 0;
`;

const BoxGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
`;

const BoxCard = styled.div`
  border: 1px solid ${props => props.theme.border};
  border-radius: 6px;
  padding: 16px;
  background: ${props => props.theme.background};
`;

const BoxHeader = styled.div`
  display: flex;
  justify-content: between;
  align-items: center;
  margin-bottom: 12px;
`;

const BoxTitle = styled.h3`
  color: ${props => props.theme.textPrimary};
  font-size: 16px;
  font-weight: 600;
  margin: 0;
`;

const BoxInfo = styled.div`
  color: ${props => props.theme.textSecondary};
  font-size: 14px;
  margin-bottom: 8px;
`;

const OperationsTable = styled.table`
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
`;

const TableHeader = styled.th`
  background: ${props => props.theme.backgroundSecondary};
  color: ${props => props.theme.textPrimary};
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid ${props => props.theme.border};
  font-weight: 600;
  font-size: 14px;
`;

const TableRow = styled.tr`
  &:nth-child(even) {
    background: ${props => props.theme.backgroundSecondary};
  }
`;

const TableCell = styled.td`
  color: ${props => props.theme.textPrimary};
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.border};
  font-size: 14px;
`;

const SuccessIndicator = styled.span`
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  background-color: ${props => props.success ? '#E8F5E8' : '#FFEBEE'};
  color: ${props => props.success ? '#2E7D32' : '#C62828'};
`;

const TimeRangeSelector = styled.select`
  background: ${props => props.theme.background};
  border: 1px solid ${props => props.theme.border};
  color: ${props => props.theme.textPrimary};
  padding: 8px 12px;
  border-radius: 4px;
  margin-bottom: 16px;
`;

const LoadingMessage = styled.div`
  color: ${props => props.theme.textSecondary};
  text-align: center;
  padding: 40px;
  font-size: 16px;
`;

const ErrorMessage = styled.div`
  color: #F44336;
  background: #FFEBEE;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 16px;
`;

const RefreshButton = styled.button`
  background: ${props => props.theme.primary};
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  margin-left: auto;
  
  &:hover {
    opacity: 0.9;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const ModbusDashboard = () => {
  const theme = getTheme('light');
  const [dashboardData, setDashboardData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [timeRange, setTimeRange] = useState('24h');

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get(`/admin/modbus/dashboard?time_range=${timeRange}`);
      setDashboardData(response.data.data);
    } catch (err) {
      setError('Ошибка загрузки данных дашборда: ' + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
  }, [timeRange]);

  useEffect(() => {
    const interval = setInterval(fetchDashboardData, 30000); // Обновление каждые 30 секунд
    return () => clearInterval(interval);
  }, [timeRange]);

  const formatTime = (timestamp) => {
    return new Date(timestamp).toLocaleString('ru-RU');
  };

  const formatSuccessRate = (rate) => {
    return `${rate.toFixed(1)}%`;
  };

  if (loading && !dashboardData) {
    return (
      <Container>
        <LoadingMessage theme={theme}>
          Загрузка данных мониторинга Modbus...
        </LoadingMessage>
      </Container>
    );
  }

  if (error) {
    return (
      <Container>
        <ErrorMessage>{error}</ErrorMessage>
        <RefreshButton theme={theme} onClick={fetchDashboardData}>
          Попробовать снова
        </RefreshButton>
      </Container>
    );
  }

  if (!dashboardData) {
    return (
      <Container>
        <LoadingMessage theme={theme}>
          Нет данных для отображения
        </LoadingMessage>
      </Container>
    );
  }

  const { 
    overview = {}, 
    box_statuses = [], 
    recent_operations = [], 
    error_stats = {} 
  } = dashboardData || {};

  return (
    <Container>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <Title theme={theme}>Мониторинг Modbus</Title>
        <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
          <TimeRangeSelector 
            theme={theme} 
            value={timeRange} 
            onChange={(e) => setTimeRange(e.target.value)}
          >
            <option value="1h">Последний час</option>
            <option value="24h">Последние 24 часа</option>
            <option value="7d">Последние 7 дней</option>
            <option value="30d">Последние 30 дней</option>
          </TimeRangeSelector>
          <RefreshButton theme={theme} onClick={fetchDashboardData} disabled={loading}>
            {loading ? 'Обновление...' : 'Обновить'}
          </RefreshButton>
        </div>
      </div>

      {/* Общая статистика */}
      <StatsGrid>
        <StatCard theme={theme}>
          <StatTitle theme={theme}>Всего боксов</StatTitle>
          <StatValue theme={theme}>{overview.total_boxes || 0}</StatValue>
          <StatSubtext theme={theme}>с настроенным Modbus</StatSubtext>
        </StatCard>
        
        <StatCard theme={theme}>
          <StatTitle theme={theme}>Подключенные</StatTitle>
          <StatValue theme={theme} style={{ color: '#4CAF50' }}>
            {overview.connected_boxes || 0}
          </StatValue>
          <StatSubtext theme={theme}>боксов онлайн</StatSubtext>
        </StatCard>
        
        <StatCard theme={theme}>
          <StatTitle theme={theme}>Отключенные</StatTitle>
          <StatValue theme={theme} style={{ color: '#F44336' }}>
            {overview.disconnected_boxes || 0}
          </StatValue>
          <StatSubtext theme={theme}>боксов офлайн</StatSubtext>
        </StatCard>
        
        <StatCard theme={theme}>
          <StatTitle theme={theme}>Операции</StatTitle>
          <StatValue theme={theme}>{overview.operations_last_24h || 0}</StatValue>
          <StatSubtext theme={theme}>за выбранный период</StatSubtext>
        </StatCard>
        
        <StatCard theme={theme}>
          <StatTitle theme={theme}>Успешность</StatTitle>
          <StatValue theme={theme} style={{ color: (overview.success_rate_last_24h || 0) >= 95 ? '#4CAF50' : '#FF9800' }}>
            {formatSuccessRate(overview.success_rate_last_24h || 0)}
          </StatValue>
          <StatSubtext theme={theme}>успешных операций</StatSubtext>
        </StatCard>
      </StatsGrid>

      {/* Статус боксов */}
      <Section theme={theme}>
        <SectionTitle theme={theme}>Статус боксов</SectionTitle>
        <BoxGrid>
          {box_statuses && box_statuses.length > 0 ? box_statuses.map(box => (
            <BoxCard key={box.box_id} theme={theme}>
              <BoxHeader>
                <BoxTitle theme={theme}>
                  <StatusIndicator connected={box.connected} />
                  Бокс #{box.box_number}
                </BoxTitle>
              </BoxHeader>
              
              <BoxInfo theme={theme}>
                <strong>Статус:</strong> {box.connected ? 'Подключен' : 'Отключен'}
              </BoxInfo>
              
              {box.last_seen && (
                <BoxInfo theme={theme}>
                  <strong>Последняя активность:</strong> {formatTime(box.last_seen)}
                </BoxInfo>
              )}
              
              {box.last_error && (
                <BoxInfo theme={theme} style={{ color: '#F44336' }}>
                  <strong>Последняя ошибка:</strong> {box.last_error}
                </BoxInfo>
              )}
              
              <BoxInfo theme={theme}>
                <strong>Регистр света:</strong> {box.light_coil_register || 'Не настроен'}
              </BoxInfo>
              
              {box.chemistry_coil_register && (
                <BoxInfo theme={theme}>
                  <strong>Регистр химии:</strong> {box.chemistry_coil_register}
                </BoxInfo>
              )}
              
              <BoxInfo theme={theme}>
                <strong>Операций:</strong> {box.operations_last_24h} 
                ({formatSuccessRate(box.success_rate_last_24h)} успешных)
              </BoxInfo>
            </BoxCard>
          )) : (
            <div style={{ textAlign: 'center', color: '#666', padding: '40px' }}>
              Нет боксов с настроенным Modbus
            </div>
          )}
        </BoxGrid>
      </Section>

      {/* Последние операции */}
      <Section theme={theme}>
        <SectionTitle theme={theme}>Последние операции</SectionTitle>
        <OperationsTable>
          <thead>
            <tr>
              <TableHeader theme={theme}>Время</TableHeader>
              <TableHeader theme={theme}>Бокс</TableHeader>
              <TableHeader theme={theme}>Операция</TableHeader>
              <TableHeader theme={theme}>Регистр</TableHeader>
              <TableHeader theme={theme}>Значение</TableHeader>
              <TableHeader theme={theme}>Статус</TableHeader>
            </tr>
          </thead>
          <tbody>
            {recent_operations && recent_operations.length > 0 ? recent_operations.map(op => (
              <TableRow key={op.id}>
                <TableCell theme={theme}>{formatTime(op.created_at)}</TableCell>
                <TableCell theme={theme}>{op.box_id}</TableCell>
                <TableCell theme={theme}>{op.operation}</TableCell>
                <TableCell theme={theme}>{op.register}</TableCell>
                <TableCell theme={theme}>{op.value ? 'ВКЛ' : 'ВЫКЛ'}</TableCell>
                <TableCell theme={theme}>
                  <SuccessIndicator success={op.success}>
                    {op.success ? 'Успешно' : 'Ошибка'}
                  </SuccessIndicator>
                  {!op.success && op.error && (
                    <div style={{ fontSize: '12px', color: '#F44336', marginTop: '4px' }}>
                      {op.error}
                    </div>
                  )}
                </TableCell>
              </TableRow>
            )) : (
              <TableRow>
                <TableCell theme={theme} colSpan="6" style={{ textAlign: 'center', color: '#666', padding: '40px' }}>
                  Нет операций для отображения
                </TableCell>
              </TableRow>
            )}
          </tbody>
        </OperationsTable>
      </Section>

      {/* Статистика ошибок */}
      {Object.keys(error_stats).length > 0 && (
        <Section theme={theme}>
          <SectionTitle theme={theme}>Типы ошибок</SectionTitle>
          {Object.entries(error_stats).map(([error, count]) => (
            <BoxInfo key={error} theme={theme}>
              <strong>{error}:</strong> {count} раз
            </BoxInfo>
          ))}
        </Section>
      )}
    </Container>
  );
};

export default ModbusDashboard;
