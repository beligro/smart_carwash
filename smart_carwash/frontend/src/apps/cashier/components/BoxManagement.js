import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import usePolling from '../../../shared/hooks/usePolling';

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

const ModalOverlay = styled.div`
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
`;

const ModalCard = styled.div`
  background: ${props => props.theme.cardBackground};
  border-radius: 10px;
  padding: 24px;
  width: 100%;
  max-width: 460px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
`;

const ModalTitle = styled.h3`
  margin: 0 0 16px 0;
  color: ${props => props.theme.textColor};
`;

const ModalLabel = styled.label`
  display: block;
  font-size: 0.9rem;
  color: ${props => props.theme.textColorSecondary};
  margin-bottom: 6px;
`;

const ModalSelect = styled.select`
  width: 100%;
  padding: 10px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 6px;
  background: ${props => props.theme.cardBackground};
  color: ${props => props.theme.textColor};
  font-size: 1rem;
  margin-bottom: 16px;
`;

const ModalTextArea = styled.textarea`
  width: 100%;
  min-height: 72px;
  padding: 10px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 6px;
  background: ${props => props.theme.cardBackground};
  color: ${props => props.theme.textColor};
  font-size: 1rem;
  resize: vertical;
  margin-bottom: 20px;
  box-sizing: border-box;
`;

const ModalActions = styled.div`
  display: flex;
  gap: 12px;
  justify-content: flex-end;
`;

const ModalButton = styled.button`
  padding: 10px 18px;
  border: none;
  border-radius: 6px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;

  &:disabled { opacity: 0.5; cursor: not-allowed; }

  &.primary { background: #dc3545; color: #fff; }
  &.primary:hover:not(:disabled) { background: #c82333; }
  &.secondary { background: ${props => props.theme.borderColor || '#e0e0e0'}; color: ${props => props.theme.textColor}; }
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
          onClick={() => onSetMaintenance(box)}
          disabled={actionLoading[box.id]}
        >
          {actionLoading[box.id] ? 'Переводим...' : 'Перевести на сервис'}
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
  // Модалка постановки в сервис: выбор симптома + комментарий
  const [maintBox, setMaintBox] = useState(null); // { id, number }
  const [symptoms, setSymptoms] = useState([]);
  const [symptomId, setSymptomId] = useState('');
  const [maintComment, setMaintComment] = useState('');
  const [maintSubmitting, setMaintSubmitting] = useState(false);

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

  // Тихий поллинг без показа загрузки
  const pollBoxes = async () => {
    try {
      const response = await ApiService.getCashierWashBoxes(filters);
      const newBoxes = response.wash_boxes || [];
      
      // Обновляем только если данные изменились
      setBoxes(prevBoxes => {
        if (JSON.stringify(prevBoxes) !== JSON.stringify(newBoxes)) {
          return newBoxes;
        }
        return prevBoxes;
      });
    } catch (error) {
      console.error('Ошибка поллинга боксов:', error);
      // Не показываем ошибку при поллинге, чтобы не мешать пользователю
    }
  };

  // Открытие модалки постановки в сервис: подгружаем симптомы по номеру бокса
  const handleSetMaintenance = async (box) => {
    setMaintBox(box);
    setSymptomId('');
    setMaintComment('');
    setSymptoms([]);
    try {
      const resp = await ApiService.getServiceSymptoms(box.number);
      setSymptoms(resp.symptoms || []);
    } catch (error) {
      console.error('Ошибка загрузки симптомов:', error);
      // Модалку не закрываем: можно поставить в сервис без выбора симптома
    }
  };

  const closeMaintModal = () => {
    setMaintBox(null);
    setSymptoms([]);
    setSymptomId('');
    setMaintComment('');
  };

  // Подтверждение: открываем наряд (бокс -> сервис + симптом + комментарий)
  const submitMaintenance = async () => {
    if (!maintBox) return;
    const boxId = maintBox.id;
    setMaintSubmitting(true);
    setActionLoading(prev => ({ ...prev, [boxId]: true }));
    try {
      await ApiService.createServiceTicket({
        box_id: boxId,
        symptom_id: symptomId ? Number(symptomId) : null,
        comment: maintComment,
      });
      closeMaintModal();
      await loadBoxes();
    } catch (error) {
      console.error('Ошибка перевода бокса на сервис:', error);
      setError('Ошибка перевода бокса на сервис: ' + (error.response?.data?.error || error.message));
    } finally {
      setMaintSubmitting(false);
      setActionLoading(prev => ({ ...prev, [boxId]: false }));
    }
  };

  const handleFilterChange = (filterType, value) => {
    setFilters(prev => ({
      ...prev,
      [filterType]: value === '' ? undefined : value
    }));
  };

  // Поллинг для автоматического обновления данных каждые 5 секунд (тихий)
  usePolling(pollBoxes, 5000, true, [filters]);

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

      {maintBox && (
        <ModalOverlay onClick={maintSubmitting ? undefined : closeMaintModal}>
          <ModalCard theme={theme} onClick={(e) => e.stopPropagation()}>
            <ModalTitle theme={theme}>Бокс #{maintBox.number} → сервис</ModalTitle>

            <ModalLabel theme={theme}>Что случилось?</ModalLabel>
            <ModalSelect
              theme={theme}
              value={symptomId}
              onChange={(e) => setSymptomId(e.target.value)}
            >
              <option value="">— выберите симптом —</option>
              {symptoms.map(s => (
                <option key={s.id} value={s.id}>{s.group_name}: {s.name}</option>
              ))}
            </ModalSelect>

            <ModalLabel theme={theme}>Комментарий (необязательно)</ModalLabel>
            <ModalTextArea
              theme={theme}
              value={maintComment}
              onChange={(e) => setMaintComment(e.target.value)}
              placeholder="Детали для мастера"
            />

            <ModalActions>
              <ModalButton
                className="secondary"
                theme={theme}
                onClick={closeMaintModal}
                disabled={maintSubmitting}
              >
                Отмена
              </ModalButton>
              <ModalButton
                className="primary"
                onClick={submitMaintenance}
                disabled={maintSubmitting}
              >
                {maintSubmitting ? 'Переводим...' : 'Перевести на сервис'}
              </ModalButton>
            </ModalActions>
          </ModalCard>
        </ModalOverlay>
      )}
    </Container>
  );
};

export default BoxManagement;


