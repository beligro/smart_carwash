import React from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  margin-top: 20px;
`;

const StatisticsTitle = styled.h3`
  margin: 0 0 15px 0;
  color: ${props => props.theme.textColor};
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  gap: 10px;
`;

const StatisticsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
`;

const StatCard = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border-left: 4px solid ${props => props.isTotal ? props.theme.primaryColor : props.theme.borderColor};
`;

const StatTitle = styled.h4`
  margin: 0 0 10px 0;
  color: ${props => props.theme.textColor};
  font-size: 1rem;
  font-weight: 600;
`;

const StatValue = styled.div`
  font-size: 1.5rem;
  font-weight: bold;
  color: ${props => props.theme.primaryColor};
  margin-bottom: 5px;
`;

const StatSubValue = styled.div`
  font-size: 0.9rem;
  color: ${props => props.theme.textSecondary};
`;

const PeriodInfo = styled.div`
  font-size: 0.9rem;
  color: ${props => props.theme.textSecondary};
  margin-bottom: 15px;
  font-style: italic;
`;

const LoadingText = styled.div`
  text-align: center;
  color: ${props => props.theme.textSecondary};
  font-style: italic;
  padding: 20px;
`;

const ErrorText = styled.div`
  text-align: center;
  color: ${props => props.theme.errorColor};
  padding: 20px;
`;

const PaymentStatistics = ({ filters, onClose }) => {
  const [statistics, setStatistics] = React.useState(null);
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState(null);

  const loadStatistics = React.useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await ApiService.getPaymentStatistics(filters);
      setStatistics(response);
    } catch (err) {
      setError('Ошибка при загрузке статистики');
    } finally {
      setLoading(false);
    }
  }, [filters]);

  React.useEffect(() => {
    loadStatistics();
  }, [loadStatistics]);

  // Обновляем статистику при изменении фильтров
  React.useEffect(() => {
    loadStatistics();
  }, [filters]);

  const formatAmount = (amount) => {
    // Принудительно преобразуем в число
    const numAmount = Number(amount);
    if (!numAmount || numAmount === 0) return '0 ₽';
    const rubles = numAmount / 100;
    return `${rubles.toFixed(2)} ₽`;
  };

  const getServiceTypeName = (serviceType, withChemistry) => {
    switch (serviceType) {
      case 'wash':
        return withChemistry ? 'Мойка + химия' : 'Мойка';
      case 'air_dry':
        return 'Воздух';
      case 'vacuum':
        return 'Пылесос';
      case 'total':
        return 'Итого';
      default:
        return serviceType;
    }
  };

  if (loading) {
    return (
      <Container>
        <LoadingText>Загрузка статистики...</LoadingText>
      </Container>
    );
  }

  if (error) {
    return (
      <Container>
        <ErrorText>{error}</ErrorText>
      </Container>
    );
  }

  if (!statistics) {
    return null;
  }

  // Проверяем, есть ли данные в статистике
  if (!statistics.statistics || statistics.statistics.length === 0) {
    return (
      <Container>
        <StatisticsTitle>
          📊 Статистика платежей
          {onClose && (
            <button 
              onClick={onClose}
              style={{
                background: 'none',
                border: 'none',
                color: getTheme().textSecondary,
                cursor: 'pointer',
                fontSize: '1.2rem',
                marginLeft: 'auto'
              }}
            >
              ✕
            </button>
          )}
        </StatisticsTitle>
        
        <PeriodInfo>Период: {statistics.period}</PeriodInfo>
        <div style={{ fontSize: '0.8rem', color: getTheme().textSecondary, marginBottom: '15px', fontStyle: 'italic' }}>
          * Суммы указаны с учетом возвратов
        </div>
        
        <div style={{ textAlign: 'center', padding: '20px', color: getTheme().textSecondary }}>
          По выбранным фильтрам данных не найдено
        </div>
      </Container>
    );
  }

  return (
    <Container>
      <StatisticsTitle>
        📊 Статистика платежей
        {onClose && (
          <button 
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              color: getTheme().textSecondary,
              cursor: 'pointer',
              fontSize: '1.2rem',
              marginLeft: 'auto'
            }}
          >
            ✕
          </button>
        )}
      </StatisticsTitle>
      
      <PeriodInfo>Период: {statistics.period}</PeriodInfo>
      <div style={{ fontSize: '0.8rem', color: getTheme().textSecondary, marginBottom: '15px', fontStyle: 'italic' }}>
        * Суммы указаны с учетом возвратов
      </div>

      <StatisticsGrid>
        {statistics.statistics.map((stat, index) => {
          return (
            <StatCard key={`${stat.service_type}-${stat.with_chemistry}`} isTotal={false}>
              <StatTitle>{getServiceTypeName(stat.service_type, stat.with_chemistry)}</StatTitle>
              <StatValue>{formatAmount(stat.total_amount)}</StatValue>
              <StatSubValue>{stat.session_count} сессий</StatSubValue>
            </StatCard>
          );
        })}
        
        {statistics.total.session_count > 0 && (
          <StatCard isTotal={true}>
            <StatTitle>{getServiceTypeName(statistics.total.service_type, statistics.total.with_chemistry)}</StatTitle>
            <StatValue>{formatAmount(statistics.total.total_amount)}</StatValue>
            <StatSubValue>{statistics.total.session_count} сессий</StatSubValue>
          </StatCard>
        )}
      </StatisticsGrid>
    </Container>
  );
};

export default PaymentStatistics; 