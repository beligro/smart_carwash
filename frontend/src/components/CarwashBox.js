import React from 'react';
import styled from 'styled-components';

const BoxContainer = styled.div`
  background-color: var(--tg-theme-secondary-bg-color, white);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
`;

const BoxInfo = styled.div`
  display: flex;
  flex-direction: column;
`;

const BoxName = styled.h3`
  margin: 0 0 8px 0;
  font-size: 18px;
`;

const StatusIndicator = styled.div`
  display: flex;
  align-items: center;
`;

const StatusDot = styled.div`
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-right: 8px;
  background-color: ${({ status }) => {
    switch (status) {
      case 'available':
        return 'var(--success-color, #28a745)';
      case 'occupied':
        return 'var(--warning-color, #ffc107)';
      case 'maintenance':
        return 'var(--danger-color, #dc3545)';
      default:
        return 'var(--secondary-color, #6c757d)';
    }
  }};
`;

const StatusText = styled.span`
  font-size: 14px;
  color: var(--tg-theme-hint-color, #6c757d);
`;

const getStatusText = (status) => {
  switch (status) {
    case 'available':
      return 'Доступен';
    case 'occupied':
      return 'Занят';
    case 'maintenance':
      return 'На обслуживании';
    default:
      return 'Неизвестно';
  }
};

const CarwashBox = ({ box }) => {
  return (
    <BoxContainer>
      <BoxInfo>
        <BoxName>{box.name}</BoxName>
        <StatusIndicator>
          <StatusDot status={box.status} />
          <StatusText>{getStatusText(box.status)}</StatusText>
        </StatusIndicator>
      </BoxInfo>
    </BoxContainer>
  );
};

export default CarwashBox;
