import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  padding: 20px;
`;

const Title = styled.h1`
  margin: 0 0 30px 0;
  color: ${props => props.theme.textColor};
  font-size: 2rem;
`;

const StatsContainer = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 25px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
`;

const StatsTitle = styled.h2`
  margin: 0 0 20px 0;
  color: ${props => props.theme.textColor};
  font-size: 1.5rem;
`;

const FiltersContainer = styled.div`
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  align-items: center;
  flex-wrap: wrap;
`;

const FilterGroup = styled.div`
  display: flex;
  flex-direction: column;
`;

const FilterLabel = styled.label`
  margin-bottom: 5px;
  color: ${props => props.theme.textColor};
  font-weight: 500;
  font-size: 0.9rem;
`;

const FilterInput = styled.input`
  padding: 8px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;

  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const Button = styled.button`
  padding: 8px 16px;
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background-color 0.2s;

  &:hover {
    background-color: ${props => props.theme.primaryColorHover};
  }

  &:disabled {
    background-color: #ccc;
    cursor: not-allowed;
  }
`;

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
`;

const StatCard = styled.div`
  background-color: ${props => props.theme.inputBackground};
  padding: 20px;
  border-radius: 8px;
  border: 1px solid ${props => props.theme.borderColor};
  text-align: center;
`;

const StatValue = styled.div`
  font-size: 2rem;
  font-weight: bold;
  color: ${props => props.theme.primaryColor};
  margin-bottom: 8px;
`;

const StatLabel = styled.div`
  font-size: 0.9rem;
  color: ${props => props.theme.textColorSecondary};
`;

const StatDescription = styled.div`
  font-size: 0.8rem;
  color: ${props => props.theme.textColorSecondary};
  margin-top: 8px;
`;

const LoadingSpinner = styled.div`
  text-align: center;
  padding: 40px;
  color: ${props => props.theme.textColorSecondary};
  font-size: 1.1rem;
`;

const ErrorMessage = styled.div`
  padding: 15px;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 4px;
  color: #dc2626;
  margin-bottom: 20px;
`;

const ChemistryStats = () => {
  const theme = getTheme('light');
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  const loadStats = async () => {
    setLoading(true);
    setError('');

    try {
      const filters = {};
      if (dateFrom) filters.dateFrom = dateFrom;
      if (dateTo) filters.dateTo = dateTo;

      const response = await ApiService.getChemistryStats(filters);
      setStats(response.stats);
    } catch (err) {
      setError('Ошибка при загрузке статистики: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStats();
  }, []);

  const handleFilterChange = () => {
    loadStats();
  };

  const formatPercentage = (percentage) => {
    return `${percentage.toFixed(1)}%`;
  };

  if (loading && !stats) {
    return (
      <Container>
        <Title theme={theme}>Статистика химии</Title>
        <LoadingSpinner theme={theme}>Загрузка статистики...</LoadingSpinner>
      </Container>
    );
  }

  return (
    <Container>
      <Title theme={theme}>Статистика химии</Title>

      {error && <ErrorMessage>{error}</ErrorMessage>}

      <StatsContainer theme={theme}>
        <StatsTitle theme={theme}>Фильтры</StatsTitle>
        
        <FiltersContainer>
          <FilterGroup>
            <FilterLabel theme={theme}>Дата начала</FilterLabel>
            <FilterInput
              theme={theme}
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
            />
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>Дата окончания</FilterLabel>
            <FilterInput
              theme={theme}
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
            />
          </FilterGroup>

          <Button theme={theme} onClick={handleFilterChange} disabled={loading}>
            {loading ? 'Загрузка...' : 'Применить фильтры'}
          </Button>
        </FiltersContainer>

        {stats && (
          <>
            <StatsTitle theme={theme}>Статистика за период: {stats.period}</StatsTitle>
            
            <StatsGrid>
              <StatCard theme={theme}>
                <StatValue theme={theme}>{stats.total_sessions_with_chemistry}</StatValue>
                <StatLabel theme={theme}>Сессий с химией</StatLabel>
                <StatDescription theme={theme}>
                  Общее количество сессий, где была оплачена химия
                </StatDescription>
              </StatCard>

              <StatCard theme={theme}>
                <StatValue theme={theme}>{stats.total_chemistry_enabled}</StatValue>
                <StatLabel theme={theme}>Химия включена</StatLabel>
                <StatDescription theme={theme}>
                  Количество сессий, где химия была фактически включена
                </StatDescription>
              </StatCard>

              <StatCard theme={theme}>
                <StatValue theme={theme}>{formatPercentage(stats.usage_percentage)}</StatValue>
                <StatLabel theme={theme}>Процент использования</StatLabel>
                <StatDescription theme={theme}>
                  Процент сессий с химией, где химия была включена
                </StatDescription>
              </StatCard>
            </StatsGrid>

            <div style={{ 
              padding: '15px', 
              backgroundColor: '#f0f9ff', 
              borderRadius: '4px',
              border: '1px solid #bae6fd'
            }}>
              <h4 style={{ margin: '0 0 10px 0', color: '#0369a1' }}>Анализ</h4>
              <p style={{ margin: '0', color: '#0369a1', fontSize: '0.9rem' }}>
                {stats.usage_percentage >= 80 ? 
                  'Отличный показатель! Большинство клиентов используют химию после оплаты.' :
                  stats.usage_percentage >= 50 ?
                  'Средний показатель. Возможно, стоит улучшить информирование клиентов о возможности включения химии.' :
                  'Низкий показатель. Рекомендуется проверить удобство интерфейса и информирование клиентов.'
                }
              </p>
            </div>
          </>
        )}
      </StatsContainer>
    </Container>
  );
};

export default ChemistryStats; 