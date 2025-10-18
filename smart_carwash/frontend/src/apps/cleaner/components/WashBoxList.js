import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  margin-top: 16px;
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
`;

const Th = styled.th`
  text-align: left;
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  font-weight: 600;
  color: ${props => props.theme.textColor};
`;

const Td = styled.td`
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  color: ${props => props.theme.textColor};
`;

const StatusBadge = styled.span`
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  background-color: ${props => {
    switch (props.status) {
      case 'free': return '#28a745';
      case 'busy': return '#dc3545';
      case 'reserved': return '#ffc107';
      case 'maintenance': return '#6c757d';
      case 'cleaning': return '#17a2b8';
      default: return '#6c757d';
    }
  }};
  color: white;
`;

const ActionButton = styled.button`
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  margin-right: 8px;

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &.reserve {
    background-color: #007bff;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #0056b3;
    }
  }

  &.start {
    background-color: #28a745;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #218838;
    }
  }

  &.cancel {
    background-color: #dc3545;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #c82333;
    }
  }

  &.complete {
    background-color: #17a2b8;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #138496;
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

/**
 * Компонент списка боксов для уборщика
 * @param {Object} props - Свойства компонента
 * @param {Function} props.onCleaningAction - Callback для обновления таймера после действий с уборкой
 * @returns {React.ReactNode} - Компонент списка боксов
 */
const WashBoxList = ({ onCleaningAction }) => {
  const theme = getTheme('light');
  const [washBoxes, setWashBoxes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState({});
  const [activeCleaningBox, setActiveCleaningBox] = useState(null);

  useEffect(() => {
    loadWashBoxes();
  }, []);

  const loadWashBoxes = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await ApiService.getCleanerWashBoxes();
      const boxes = response.wash_boxes || [];
      setWashBoxes(boxes);
      
      // Определяем активную уборку
      const cleaningBox = boxes.find(box => box.status === 'cleaning');
      setActiveCleaningBox(cleaningBox || null);
    } catch (err) {
      setError('Ошибка при загрузке списка боксов');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Функция для извлечения конкретного сообщения об ошибке
  const getErrorMessage = (error, defaultMessage) => {
    if (error?.response?.data?.error) {
      return error.response.data.error;
    }
    return defaultMessage;
  };

  const handleReserveCleaning = async (washBoxId) => {
    setActionLoading(prev => ({ ...prev, [washBoxId]: true }));
    
    try {
      await ApiService.reserveCleaning(washBoxId);
      await loadWashBoxes();
      // Обновляем таймер уборки
      if (onCleaningAction) {
        onCleaningAction();
      }
    } catch (error) {
      setError(getErrorMessage(error, 'Ошибка при резервировании уборки'));
      console.error(error);
    } finally {
      setActionLoading(prev => ({ ...prev, [washBoxId]: false }));
    }
  };

  const handleStartCleaning = async (washBoxId) => {
    setActionLoading(prev => ({ ...prev, [washBoxId]: true }));
    
    try {
      await ApiService.startCleaning(washBoxId);
      await loadWashBoxes();
      // Обновляем таймер уборки
      if (onCleaningAction) {
        onCleaningAction();
      }
    } catch (error) {
      setError(getErrorMessage(error, 'Ошибка при начале уборки'));
      console.error(error);
    } finally {
      setActionLoading(prev => ({ ...prev, [washBoxId]: false }));
    }
  };

  const handleCancelCleaning = async (washBoxId) => {
    setActionLoading(prev => ({ ...prev, [washBoxId]: true }));
    
    try {
      await ApiService.cancelCleaning(washBoxId);
      await loadWashBoxes();
      // Обновляем таймер уборки
      if (onCleaningAction) {
        onCleaningAction();
      }
    } catch (error) {
      setError(getErrorMessage(error, 'Ошибка при отмене уборки'));
      console.error(error);
    } finally {
      setActionLoading(prev => ({ ...prev, [washBoxId]: false }));
    }
  };

  const handleCompleteCleaning = async (washBoxId) => {
    setActionLoading(prev => ({ ...prev, [washBoxId]: true }));
    
    try {
      await ApiService.completeCleaning(washBoxId);
      await loadWashBoxes();
      // Обновляем таймер уборки
      if (onCleaningAction) {
        onCleaningAction();
      }
    } catch (error) {
      setError(getErrorMessage(error, 'Ошибка при завершении уборки'));
      console.error(error);
    } finally {
      setActionLoading(prev => ({ ...prev, [washBoxId]: false }));
    }
  };

  const getStatusText = (status) => {
    switch (status) {
      case 'free': return 'Свободен';
      case 'busy': return 'Занят';
      case 'reserved': return 'Зарезервирован';
      case 'maintenance': return 'На обслуживании';
      case 'cleaning': return 'На уборке';
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

  const renderActions = (washBox) => {
    const isLoading = actionLoading[washBox.id];

    // Если есть активная уборка и это не текущий бокс, скрываем кнопки
    if (activeCleaningBox && activeCleaningBox.id !== washBox.id) {
      return (
        <span style={{ color: '#6c757d', fontStyle: 'italic' }}>
          Убирает бокс №{activeCleaningBox.number}
        </span>
      );
    }

    switch (washBox.status) {
      case 'free':
        return (
          <ActionButton
            className="start"
            onClick={() => handleStartCleaning(washBox.id)}
            disabled={isLoading}
          >
            {isLoading ? 'Начинаем...' : 'Начать уборку'}
          </ActionButton>
        );
      
      case 'busy':
        return (
          <ActionButton
            className="reserve"
            onClick={() => handleReserveCleaning(washBox.id)}
            disabled={isLoading}
          >
            {isLoading ? 'Резервируем...' : 'Зарезервировать уборку'}
          </ActionButton>
        );
      
      case 'cleaning':
        return (
          <ActionButton
            className="complete"
            onClick={() => handleCompleteCleaning(washBox.id)}
            disabled={isLoading}
          >
            {isLoading ? 'Завершаем...' : 'Завершить уборку'}
          </ActionButton>
        );
      
      default:
        if (washBox.cleaning_reserved_by) {
          return (
            <ActionButton
              className="cancel"
              onClick={() => handleCancelCleaning(washBox.id)}
              disabled={isLoading}
            >
              {isLoading ? 'Отменяем...' : 'Отменить резерв'}
            </ActionButton>
          );
        }
        return null;
    }
  };

  if (loading) {
    return <LoadingSpinner theme={theme}>Загрузка боксов...</LoadingSpinner>;
  }

  return (
    <Container>
      {error && <ErrorMessage>{error}</ErrorMessage>}
      
      <Table>
        <thead>
          <tr>
            <Th theme={theme}>Номер</Th>
            <Th theme={theme}>Тип услуги</Th>
            <Th theme={theme}>Статус</Th>
            <Th theme={theme}>Химия</Th>
            <Th theme={theme}>Действия</Th>
          </tr>
        </thead>
        <tbody>
          {washBoxes.map(washBox => (
            <tr key={washBox.id}>
              <Td theme={theme}>{washBox.number}</Td>
              <Td theme={theme}>{getServiceTypeText(washBox.service_type)}</Td>
              <Td theme={theme}>
                <StatusBadge status={washBox.status}>
                  {getStatusText(washBox.status)}
                </StatusBadge>
              </Td>
              <Td theme={theme}>
                {washBox.chemistry_enabled ? 'Доступна' : 'Недоступна'}
              </Td>
              <Td theme={theme}>
                {renderActions(washBox)}
              </Td>
            </tr>
          ))}
        </tbody>
      </Table>
    </Container>
  );
};

export default WashBoxList;

