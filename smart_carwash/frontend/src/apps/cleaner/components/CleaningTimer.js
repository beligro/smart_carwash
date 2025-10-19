import React, { useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  margin-top: 16px;
`;

const TimerCard = styled.div`
  background-color: ${props => props.theme.backgroundColor};
  border: 2px solid #17a2b8;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const BoxInfo = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
`;

const BoxNumber = styled.span`
  font-size: 1.2rem;
  font-weight: 600;
  color: ${props => props.theme.textColor};
`;

const Timer = styled.div`
  font-size: 1.5rem;
  font-weight: 700;
  color: #17a2b8;
  font-family: 'Courier New', monospace;
`;

const StatusText = styled.div`
  color: ${props => props.theme.textColorSecondary};
  font-size: 0.9rem;
`;

const NoActiveCleaning = styled.div`
  text-align: center;
  padding: 40px;
  color: ${props => props.theme.textColorSecondary};
  font-style: italic;
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
 * Компонент таймера уборки
 * @param {React.Ref} ref - Ref для передачи методов родительскому компоненту
 * @returns {React.ReactNode} - Компонент таймера
 */
const CleaningTimer = forwardRef((props, ref) => {
  const theme = getTheme('light');
  const [cleaningBoxes, setCleaningBoxes] = useState([]);
  const [error, setError] = useState(null);
  const [timeouts, setTimeouts] = useState({});
  const [cleaningTimeout, setCleaningTimeout] = useState(3); // По умолчанию 3 минуты

  useEffect(() => {
    loadCleaningTimeout();
    loadCleaningBoxes();
    
    // Обновляем данные каждые 30 секунд
    const dataInterval = setInterval(() => {
      loadCleaningTimeout();
      loadCleaningBoxes();
    }, 30000);
    
    // Обновляем UI каждую секунду для плавного таймера
    const timerInterval = setInterval(() => {
      setCleaningBoxes(prev => [...prev]); // Принудительно обновляем компонент
    }, 1000);
    
    return () => {
      clearInterval(dataInterval);
      clearInterval(timerInterval);
      // Очищаем все таймауты при размонтировании
      Object.values(timeouts).forEach(clearTimeout);
    };
  }, []);

  const loadCleaningTimeout = async () => {
    try {
      const response = await ApiService.getCleaningTimeout();
      setCleaningTimeout(response.timeout_minutes || 3);
    } catch (err) {
      console.warn('Не удалось загрузить время уборки, используем значение по умолчанию:', err);
      setCleaningTimeout(3);
    }
  };

  const loadCleaningBoxes = async () => {
    try {
      setError(null);
      const response = await ApiService.getCleanerWashBoxes();
      // API возвращает объект с полем wash_boxes, извлекаем массив
      const allBoxes = response?.wash_boxes || [];
      // Фильтруем только боксы в статусе cleaning
      const cleaningBoxes = allBoxes.filter(box => box.status === 'cleaning');
      setCleaningBoxes(cleaningBoxes);
    } catch (err) {
      setError('Ошибка при загрузке информации об уборке');
      console.error(err);
    }
  };

  // Функция для принудительного обновления данных (будет вызвана из родительского компонента)
  const refreshData = () => {
    loadCleaningTimeout();
    loadCleaningBoxes();
  };

  // Передаем методы родительскому компоненту через ref
  useImperativeHandle(ref, () => ({
    refreshData
  }));

  const formatTime = (minutes, seconds) => {
    const mins = Math.max(0, Math.floor(minutes));
    const secs = Math.max(0, Math.floor(seconds));
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const getTimeRemaining = (startedAt) => {
    if (!startedAt) return { minutes: 0, seconds: 0 };
    
    const startTime = new Date(startedAt);
    const now = new Date();
    const elapsedMs = now - startTime;
    const totalMs = cleaningTimeout * 60 * 1000;
    const remainingMs = Math.max(0, totalMs - elapsedMs);
    
    const minutes = remainingMs / (60 * 1000);
    const seconds = (remainingMs % (60 * 1000)) / 1000;
    
    return { minutes, seconds };
  };

  const renderCleaningBox = (washBox) => {
    const { minutes, seconds } = getTimeRemaining(washBox.cleaning_started_at);
    const isExpired = minutes === 0 && seconds === 0;
    const isWarning = minutes < 1 && !isExpired; // Предупреждение за минуту до окончания

    return (
      <TimerCard key={washBox.id} theme={theme} style={{ 
        borderColor: isExpired ? '#dc3545' : isWarning ? '#ffc107' : '#17a2b8' 
      }}>
        <BoxInfo>
          <BoxNumber theme={theme}>Бокс №{washBox.number}</BoxNumber>
          <StatusText theme={theme}>
            {isExpired ? 'Время истекло' : isWarning ? 'Заканчивается через минуту' : 'Идет уборка'}
          </StatusText>
        </BoxInfo>
        <Timer style={{ 
          color: isExpired ? '#dc3545' : isWarning ? '#ffc107' : '#17a2b8' 
        }}>
          {isExpired ? '00:00' : formatTime(minutes, seconds)}
        </Timer>
      </TimerCard>
    );
  };

  if (error) {
    return <ErrorMessage>{error}</ErrorMessage>;
  }

  if (cleaningBoxes.length === 0) {
    return (
      <Container>
        <NoActiveCleaning theme={theme}>
          В данный момент уборка не проводится
        </NoActiveCleaning>
      </Container>
    );
  }

  return (
    <Container>
      {cleaningBoxes.map(renderCleaningBox)}
    </Container>
  );
});

CleaningTimer.displayName = 'CleaningTimer';

export default CleaningTimer;

