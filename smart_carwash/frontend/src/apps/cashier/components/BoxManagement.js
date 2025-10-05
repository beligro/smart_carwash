import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  padding: 20px;
`;

const BoxCard = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border-left: 4px solid ${props => {
    switch (props.status) {
      case 'free': return '#28a745';
      case 'reserved': return '#007bff';
      case 'busy': return '#ffc107';
      case 'maintenance': return '#dc3545';
      default: return '#6c757d';
    }
  }};
`;

const BoxHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
`;

const BoxInfo = styled.div`
  flex: 1;
`;

const BoxNumber = styled.div`
  font-weight: 600;
  font-size: 1.2rem;
  color: ${props => props.theme.textColor};
`;

const BoxStatus = styled.div`
  font-size: 0.9rem;
  color: ${props => props.theme.textColorSecondary};
  margin-top: 4px;
`;

const BoxDetails = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
`;

const DetailItem = styled.div`
  display: flex;
  flex-direction: column;
`;

const DetailLabel = styled.span`
  font-size: 0.8rem;
  color: ${props => props.theme.textColorSecondary};
  margin-bottom: 4px;
`;

const DetailValue = styled.span`
  font-size: 0.9rem;
  color: ${props => props.theme.textColor};
  font-weight: 500;
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

  &.maintenance {
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

const EmptyMessage = styled.div`
  text-align: center;
  padding: 40px;
  color: ${props => props.theme.textColorSecondary};
  font-size: 1.1rem;
`;

const FilterContainer = styled.div`
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
`;

const FilterSelect = styled.select`
  padding: 8px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.cardBackground};
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;
`;

const FilterLabel = styled.label`
  font-size: 0.9rem;
  color: ${props => props.theme.textColor};
  margin-right: 8px;
`;

const getStatusText = (status) => {
  switch (status) {
    case 'free': return 'Свободен';
    case 'reserved': return 'Зарезервирован';
    case 'busy': return 'Занят';
    case 'maintenance': return 'На обслуживании';
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

const getChemistryText = (chemistryEnabled) => {
  return chemistryEnabled ? 'Да' : 'Нет';
};

/**
 * Компонент для отображения одного бокса
 */
const BoxCardComponent = ({ box, onSetMaintenance, actionLoading, theme }) => {
  return (
    <BoxCard theme={theme} status={box.status}>
      <BoxHeader>
        <BoxInfo>
          <BoxNumber theme={theme}>Бокс #{box.number}</BoxNumber>
          <BoxStatus theme={theme}>
            {getStatusText(box.status)}
          </BoxStatus>
        </BoxInfo>
      </BoxHeader>

      <BoxDetails>
        <DetailItem>
          <DetailLabel theme={theme}>Тип услуги</DetailLabel>
          <DetailValue theme={theme}>
            {getServiceTypeText(box.service_type)}
          </DetailValue>
        </DetailItem>

        <DetailItem>
          <DetailLabel theme={theme}>Химия</DetailLabel>
          <DetailValue theme={theme}>
            {getChemistryText(box.chemistry_enabled)}
          </DetailValue>
        </DetailItem>
      </BoxDetails>

      {box.status === 'free' && (
        <ActionButton
          className="maintenance"
          onClick={() => onSetMaintenance(box.id)}
          disabled={actionLoading[box.id]}
        >
          {actionLoading[box.id] ? 'Переводим...' : 'Перевести в обслуживание'}
        </ActionButton>
      )}
    </BoxCard>
  );
};

/**
 * Компонент для управления боксами кассира
 */
const BoxManagement = () => {
  const theme = getTheme('light');
  const [boxes, setBoxes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState({});
  const [filters, setFilters] = useState({
    status: '',
    serviceType: ''
  });

  useEffect(() => {
    loadBoxes();
  }, [filters]);

  const loadBoxes = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await ApiService.getCashierWashBoxes(filters);
      setBoxes(response.wash_boxes || []);
    } catch (error) {
      console.error('Ошибка загрузки боксов:', error);
      setError('Ошибка загрузки боксов');
    } finally {
      setLoading(false);
    }
  };

  const handleSetMaintenance = async (boxId) => {
    if (!window.confirm('Вы уверены, что хотите перевести этот бокс в режим обслуживания?')) {
      return;
    }

    setActionLoading(prev => ({ ...prev, [boxId]: true }));
    
    try {
      await ApiService.setCashierMaintenance(boxId);
      await loadBoxes(); // Перезагружаем список
    } catch (error) {
      console.error('Ошибка перевода бокса в обслуживание:', error);
      setError('Ошибка перевода бокса в обслуживание: ' + (error.response?.data?.error || error.message));
    } finally {
      setActionLoading(prev => ({ ...prev, [boxId]: false }));
    }
  };

  const handleFilterChange = (filterType, value) => {
    setFilters(prev => ({
      ...prev,
      [filterType]: value === '' ? undefined : value
    }));
  };

  if (loading) {
    return <LoadingSpinner theme={theme}>Загрузка боксов...</LoadingSpinner>;
  }

  return (
    <Container theme={theme}>
      {error && (
        <ErrorMessage>{error}</ErrorMessage>
      )}

      <h3>Управление боксами</h3>
      
      <FilterContainer>
        <FilterLabel theme={theme}>Статус:</FilterLabel>
        <FilterSelect 
          theme={theme}
          value={filters.status || ''}
          onChange={(e) => handleFilterChange('status', e.target.value)}
        >
          <option value="">Все статусы</option>
          <option value="free">Свободен</option>
          <option value="reserved">Зарезервирован</option>
          <option value="busy">Занят</option>
          <option value="maintenance">На обслуживании</option>
        </FilterSelect>

        <FilterLabel theme={theme}>Тип услуги:</FilterLabel>
        <FilterSelect 
          theme={theme}
          value={filters.serviceType || ''}
          onChange={(e) => handleFilterChange('serviceType', e.target.value)}
        >
          <option value="">Все типы</option>
          <option value="wash">Мойка</option>
          <option value="air_dry">Сушка</option>
          <option value="vacuum">Пылесос</option>
        </FilterSelect>
      </FilterContainer>
      
      {boxes.length === 0 ? (
        <EmptyMessage theme={theme}>
          Боксов не найдено
        </EmptyMessage>
      ) : (
        boxes.map(box => (
          <BoxCardComponent
            key={box.id}
            box={box}
            onSetMaintenance={handleSetMaintenance}
            actionLoading={actionLoading}
            theme={theme}
          />
        ))
      )}
    </Container>
  );
};

export default BoxManagement;


